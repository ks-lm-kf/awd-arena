package network

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/awd-platform/awd-arena/pkg/logger"
)

// DockerNetworkClient abstracts Docker network operations.
type DockerNetworkClient interface {
	CreateNetwork(ctx context.Context, name, subnet, gateway string, internal bool) (string, error)
	RemoveNetwork(ctx context.Context, networkID string) error
}

// NetworkManager manages Docker networks for team isolation.
type NetworkManager struct {
	client   DockerNetworkClient
	subnet   string            // e.g. "10.10.0.0/16"
	ipBase   string            // e.g. "10.10"
	teamNets map[int64]string  // teamID -> networkName
	netIDs   map[string]string // networkName -> Docker network ID
	mu       sync.Mutex
}

// NewNetworkManager creates a new network manager.
func NewNetworkManager(client DockerNetworkClient) *NetworkManager {
	return &NetworkManager{
		client:   client,
		subnet:   "10.10.0.0/16",
		ipBase:   "10.10",
		teamNets: make(map[int64]string),
		netIDs:   make(map[string]string),
	}
}

// CreateTeamNetwork creates an isolated network for a team.
// Returns the network name (used as Docker network ID for container creation).
func (n *NetworkManager) CreateTeamNetwork(ctx context.Context, teamID int64) (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if name, ok := n.teamNets[teamID]; ok {
		return name, nil
	}

	netName := fmt.Sprintf("awd-net-team%d", teamID)
	// Each team gets a /24 subnet: 10.10.<teamID>.0/24
	teamSubnet := fmt.Sprintf("%s.%d.0/24", n.ipBase, teamID%256)
	gateway := fmt.Sprintf("%s.%d.1", n.ipBase, teamID%256)

	networkID, err := n.client.CreateNetwork(ctx, netName, teamSubnet, gateway, true)
	if err != nil {
		// Network already exists (e.g. after restart) - still track it by name
		n.teamNets[teamID] = netName
		n.netIDs[netName] = netName
		logger.Info("team network already exists, re-using", "team", teamID, "network", netName)
		return netName, nil
	}

	n.teamNets[teamID] = netName
	n.netIDs[netName] = networkID
	logger.Info("team network created", "team", teamID, "network", netName, "subnet", teamSubnet)
	return netName, nil
}

// RemoveTeamNetwork removes a team's network.
func (n *NetworkManager) RemoveTeamNetwork(ctx context.Context, teamID int64) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	netName, ok := n.teamNets[teamID]
	if !ok {
		return nil
	}

	if netID, exists := n.netIDs[netName]; exists {
		_ = n.client.RemoveNetwork(ctx, netID)
		delete(n.netIDs, netName)
	}
	delete(n.teamNets, teamID)
	logger.Info("team network removed", "team", teamID)
	return nil
}

// GetTeamIP returns the IP address for a container in a team's network.
// Team containers get IPs like 10.10.<teamID>.<containerIndex+2>
func (n *NetworkManager) GetTeamIP(teamID int64, containerIndex int) string {
	return fmt.Sprintf("%s.%d.%d", n.ipBase, teamID%256, containerIndex+2)
}

// GetTeamNetwork returns the network name for a team.
func (n *NetworkManager) GetTeamNetwork(teamID int64) (string, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	name, ok := n.teamNets[teamID]
	return name, ok
}

// IsolateTeams configures iptables rules to prevent cross-team traffic.
// Docker internal networks alone are NOT sufficient — containers on different
// bridges can still communicate via the Docker host if IP forwarding is enabled.
// We add explicit iptables FORWARD DROP rules between team subnets.
func (n *NetworkManager) IsolateTeams(ctx context.Context, teamIDs []int64) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if len(teamIDs) <= 1 {
		logger.Info("team isolation skipped (0 or 1 team)", "count", len(teamIDs))
		return nil
	}

	// Build list of team subnets
	type teamSubnet struct {
		teamID int64
		subnet string // e.g. "10.10.1.0/24"
	}
	var subnets []teamSubnet
	for _, tid := range teamIDs {
		sn := fmt.Sprintf("%s.%d.0/24", n.ipBase, (tid%254)+1)
		subnets = append(subnets, teamSubnet{teamID: tid, subnet: sn})
	}

	// Drop cross-team forwarding at the iptables level
	for i, src := range subnets {
		for j, dst := range subnets {
			if i == j {
				continue // same team — allow
			}
			// iptables -C FORWARD ... checks if rule already exists (exit 0 = exists)
			checkArgs := []string{
				"-C", "FORWARD", "-s", src.subnet, "-d", dst.subnet, "-j", "DROP",
			}
			if err := exec.Command("iptables", checkArgs...).Run(); err != nil {
				// Rule doesn't exist, insert it
				insertArgs := []string{
					"-I", "FORWARD", "-s", src.subnet, "-d", dst.subnet, "-j", "DROP",
				}
				if err := exec.Command("iptables", insertArgs...).Run(); err != nil {
					logger.Warn("failed to add isolation iptables rule (may need root/capabilities)",
						"src_team", src.teamID, "dst_team", dst.teamID, "error", err)
				}
			}
		}
	}

	logger.Info("team network isolation applied via iptables", "teams", len(teamIDs))
	return nil
}

// Cleanup removes all managed networks and isolation rules.
func (n *NetworkManager) Cleanup(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.removeIsolationRulesLocked()

	for teamID, netName := range n.teamNets {
		if netID, exists := n.netIDs[netName]; exists {
			_ = n.client.RemoveNetwork(ctx, netID)
		}
		delete(n.teamNets, teamID)
		delete(n.netIDs, netName)
	}
	logger.Info("all team networks cleaned up")
	return nil
}

func (n *NetworkManager) removeIsolationRulesLocked() {
	for tid := range n.teamNets {
		sn := fmt.Sprintf("%s.%d.0/24", n.ipBase, (tid%254)+1)
		args := []string{"-D", "FORWARD", "-s", sn, "-j", "DROP"}
		_ = exec.Command("iptables", args...).Run()
	}
}

// CreateAdminNetwork creates a network that can reach all team networks.
func (n *NetworkManager) CreateAdminNetwork(ctx context.Context) (string, error) {
	adminNet := "awd-net-admin"
	adminSubnet := "10.255.0.0/24"
	gateway := "10.255.0.1"
	networkID, err := n.client.CreateNetwork(ctx, adminNet, adminSubnet, gateway, false)
	if err != nil {
		return "", fmt.Errorf("create admin network: %w", err)
	}
	n.netIDs[adminNet] = networkID
	logger.Info("admin network created", "id", networkID[:12])
	return adminNet, nil
}
