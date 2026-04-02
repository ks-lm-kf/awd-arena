package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/awd-platform/awd-arena/pkg/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// int64Ptr converts int64 to *int64
func int64Ptr(i int64) *int64 {
	return &i
}

// DockerClientImpl implements DockerClient using the Docker SDK.
type DockerClientImpl struct {
	cli *client.Client
}

// NewDockerClientImpl creates a new Docker SDK client.
func NewDockerClientImpl() (*DockerClientImpl, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &DockerClientImpl{cli: cli}, nil
}

// Close closes the Docker client.
func (d *DockerClientImpl) Close() error {
	return d.cli.Close()
}

// CreateContainer creates a Docker container.
func (d *DockerClientImpl) CreateContainer(ctx context.Context, opts CreateOptions) (string, error) {
	config := &container.Config{
		Image: opts.Image,
		Cmd:   opts.Cmd,
		Env:   opts.Env,
	}

	// 设置磁盘配额（默认 1GB）
	diskQuota := fmt.Sprintf("%dG", opts.Resources.DiskGB)
	if diskQuota == "0G" || diskQuota == "" {
		diskQuota = "1G" // 默认 1GB
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			NanoCPUs:   int64(opts.Resources.CPUCores * 1e9),
			Memory:     opts.Resources.MemoryMB * 1024 * 1024,
			MemorySwap: opts.Resources.MemoryMB * 1024 * 1024,
			PidsLimit:  int64Ptr(int64(opts.Resources.PidsLimit)),
		},
		NetworkMode:   container.NetworkMode(opts.NetworkID),
		AutoRemove:    opts.AutoRemove,
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		SecurityOpt:   []string{"no-new-privileges:true"},
		CapDrop:       []string{"ALL"},
		CapAdd:        []string{"NET_BIND_SERVICE", "CHOWN", "SETUID", "SETGID", "DAC_OVERRIDE"},
	}

	networkingConfig := &network.NetworkingConfig{}
	if opts.IPAddress != "" {
		networkingConfig.EndpointsConfig = map[string]*network.EndpointSettings{
			opts.NetworkID: {
				IPAMConfig: &network.EndpointIPAMConfig{
					IPv4Address: opts.IPAddress,
				},
			},
		}
	}

	if len(opts.Ports) > 0 {
		hostConfig.PortBindings = make(nat.PortMap)
		config.ExposedPorts = make(nat.PortSet)
		for cPort, hPort := range opts.Ports {
			port := nat.Port(fmt.Sprintf("%d/tcp", cPort))
			hostConfig.PortBindings[port] = []nat.PortBinding{{HostPort: fmt.Sprintf("%d", hPort)}}
			config.ExposedPorts[port] = struct{}{}
		}
	}

	resp, err := d.cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, opts.Name)
	if err != nil {
		return "", fmt.Errorf("container create: %w", err)
	}

	logger.Info("container created with disk quota", "id", resp.ID[:12], "disk_quota", diskQuota)
	return resp.ID, nil
}

// StartContainer starts a container.
func (d *DockerClientImpl) StartContainer(ctx context.Context, id string) error {
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

// StopContainer stops a container.
func (d *DockerClientImpl) StopContainer(ctx context.Context, id string) error {
	return d.cli.ContainerStop(ctx, id, container.StopOptions{})
}

// RemoveContainer removes a container.
func (d *DockerClientImpl) RemoveContainer(ctx context.Context, id string) error {
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
}

// PauseContainer pauses a container.
func (d *DockerClientImpl) PauseContainer(ctx context.Context, id string) error {
	return d.cli.ContainerPause(ctx, id)
}

// UnpauseContainer unpauses a container.
func (d *DockerClientImpl) UnpauseContainer(ctx context.Context, id string) error {
	return d.cli.ContainerUnpause(ctx, id)
}

// InspectContainer returns container info.
func (d *DockerClientImpl) InspectContainer(ctx context.Context, id string) (*ContainerInfo, error) {
	inspect, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	ip := ""
	for _, nw := range inspect.NetworkSettings.Networks {
		if nw.IPAddress != "" {
			ip = nw.IPAddress
			break
		}
	}
	return &ContainerInfo{
		ID:     inspect.ID,
		Name:   strings.TrimPrefix(inspect.Name, "/"),
		Status: inspect.State.Status,
		IP:     ip,
		Image:  inspect.Config.Image,
	}, nil
}

// ListContainers lists containers matching filters.
func (d *DockerClientImpl) ListContainers(ctx context.Context, filterArgs map[string]string) ([]ContainerInfo, error) {
	args := filters.NewArgs()
	for k, v := range filterArgs {
		args.Add(k, v)
	}
	containers, err := d.cli.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return nil, err
	}
	result := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		result[i] = ContainerInfo{
			ID:     c.ID,
			Name:   strings.TrimPrefix(c.Names[0], "/"),
			Status: c.State,
			Image:  c.Image,
		}
	}
	return result, nil
}

// dockerStatsJSON mirrors the Docker stats JSON structure.
type dockerStatsJSON struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     int    `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

// Stats returns container resource stats (single snapshot).
func (d *DockerClientImpl) Stats(ctx context.Context, id string) (*ContainerStats, error) {
	resp, err := d.cli.ContainerStatsOneShot(ctx, id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ds dockerStatsJSON
	if err := json.Unmarshal(body, &ds); err != nil {
		return &ContainerStats{ID: id}, nil
	}

	cpuDelta := float64(ds.CPUStats.CPUUsage.TotalUsage - ds.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(ds.CPUStats.SystemCPUUsage - ds.PreCPUStats.SystemCPUUsage)
	cpuPercent := 0.0
	if systemDelta > 0 && ds.CPUStats.OnlineCPUs > 0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(ds.CPUStats.OnlineCPUs) * 100.0
	}

	var netRx, netTx uint64
	for _, v := range ds.Networks {
		netRx += v.RxBytes
		netTx += v.TxBytes
	}

	return &ContainerStats{
		ID:      id,
		CPU:     cpuPercent,
		Memory:  ds.MemoryStats.Usage,
		Network: netRx + netTx,
	}, nil
}

// PullImage pulls a Docker image.
func (d *DockerClientImpl) PullImage(ctx context.Context, ref string) error {
	reader, err := d.cli.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	return err
}

// ListImages lists Docker images matching a filter.
func (d *DockerClientImpl) ListImages(ctx context.Context, filter string) ([]string, error) {
	args := filters.NewArgs()
	if filter != "" {
		args.Add("reference", filter)
	}
	images, err := d.cli.ImageList(ctx, image.ListOptions{Filters: args})
	if err != nil {
		return nil, err
	}
	result := make([]string, len(images))
	for i, img := range images {
		if len(img.RepoTags) > 0 {
			result[i] = img.RepoTags[0]
		} else {
			result[i] = img.ID[:12]
		}
	}
	return result, nil
}

// CreateNetwork creates a Docker network.
func (d *DockerClientImpl) CreateNetwork(ctx context.Context, name, subnet, gateway string, internal bool) (string, error) {
	ipamConfig := []network.IPAMConfig{{
		Subnet:  subnet,
		Gateway: gateway,
	}}
	resp, err := d.cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Config: ipamConfig,
		},
		Internal: internal,
		Labels: map[string]string{
			"awd-managed": "true",
			"awd-network": name,
		},
	})
	if err != nil {
		return "", err
	}
	logger.Info("network created", "name", name, "subnet", subnet, "id", resp.ID[:12])
	return resp.ID, nil
}

// RemoveNetwork removes a Docker network.
func (d *DockerClientImpl) RemoveNetwork(ctx context.Context, networkID string) error {
	return d.cli.NetworkRemove(ctx, networkID)
}

// ListNetworks lists Docker networks with a label filter.
func (d *DockerClientImpl) ListNetworks(ctx context.Context) ([]types.NetworkResource, error) {
	args := filters.NewArgs()
	args.Add("label", "awd-managed=true")
	return d.cli.NetworkList(ctx, types.NetworkListOptions{Filters: args})
}

// LogOptions holds log retrieval options.
type LogOptions struct {
	ShowStdout bool   `json:"show_stdout"`
	ShowStderr bool   `json:"show_stderr"`
	Since      string `json:"since"`
	Until      string `json:"until"`
	Tail       string `json:"tail"`
	Timestamps bool   `json:"timestamps"`
}

// Logs retrieves container logs.
func (d *DockerClientImpl) Logs(ctx context.Context, id string, opts LogOptions) (string, error) {
	options := container.LogsOptions{
		ShowStdout: opts.ShowStdout,
		ShowStderr: opts.ShowStderr,
		Since:      opts.Since,
		Until:      opts.Until,
		Tail:       opts.Tail,
		Timestamps: opts.Timestamps,
		Follow:     false,
	}
	reader, err := d.cli.ContainerLogs(ctx, id, options)
	if err != nil {
		return "", fmt.Errorf("container logs: %w", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// LogsStream streams container logs in real-time.
func (d *DockerClientImpl) LogsStream(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error) {
	options := container.LogsOptions{
		ShowStdout: opts.ShowStdout,
		ShowStderr: opts.ShowStderr,
		Since:      opts.Since,
		Tail:       opts.Tail,
		Timestamps: opts.Timestamps,
		Follow:     true,
	}
	return d.cli.ContainerLogs(ctx, id, options)
}

// ExecCreate creates an exec instance in a container.
func (d *DockerClientImpl) ExecCreate(ctx context.Context, id string, cmd []string) (string, error) {
	config := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}
	resp, err := d.cli.ContainerExecCreate(ctx, id, config)
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}
	return resp.ID, nil
}

// ExecStart starts an exec instance.
func (d *DockerClientImpl) ExecStart(ctx context.Context, execID string) (string, error) {
	resp, err := d.cli.ContainerExecAttach(ctx, execID, types.ExecStartCheck{})
	if err != nil {
		return "", fmt.Errorf("exec start: %w", err)
	}
	defer resp.Close()

	body, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// RemoveImage removes a Docker image from the host.
func (d *DockerClientImpl) RemoveImage(ctx context.Context, imageID string, force bool) error {
	_, err := d.cli.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force:         force,
		PruneChildren: false,
	})
	return err
}

// PushImage pushes a Docker image to a registry.
