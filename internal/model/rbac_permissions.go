package model

// Role represents user roles
type Role string

const (
	RoleAdmin     Role = "admin"     // Full system access
	RoleOrganizer Role = "organizer" // Competition management
	RolePlayer    Role = "player"    // Regular participant
)

// Permission represents system permissions
type Permission string

const (
	// Admin permissions
	PermManageUsers         Permission = "manage_users"
	PermManageGames         Permission = "manage_games"
	PermManageChallenges    Permission = "manage_challenges"
	PermManageInfrastructure Permission = "manage_infrastructure"
	PermManageSettings      Permission = "manage_settings"
	PermViewAllData         Permission = "view_all_data"

	// Organizer permissions
	PermCreateGame          Permission = "create_game"
	PermEditGame            Permission = "edit_game"
	PermDeleteGame          Permission = "delete_game"
	PermStartGame           Permission = "start_game"
	PermPauseGame           Permission = "pause_game"
	PermStopGame            Permission = "stop_game"
	PermViewGameStats       Permission = "view_game_stats"
	PermManageTeams         Permission = "manage_teams"
	PermCreateChallenge     Permission = "create_challenge"
	PermEditChallenge       Permission = "edit_challenge"
	PermDeleteChallenge     Permission = "delete_challenge"

	// Player permissions
	PermViewGame            Permission = "view_game"
	PermJoinGame            Permission = "join_game"
	PermSubmitFlag          Permission = "submit_flag"
	PermViewOwnStats        Permission = "view_own_stats"
	PermViewRankings        Permission = "view_rankings"
)

// RolePermission defines which permissions each role has
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermManageUsers,
		PermManageGames,
		PermManageChallenges,
		PermManageInfrastructure,
		PermManageSettings,
		PermViewAllData,
		PermCreateGame,
		PermEditGame,
		PermDeleteGame,
		PermStartGame,
		PermPauseGame,
		PermStopGame,
		PermViewGameStats,
		PermManageTeams,
		PermCreateChallenge,
		PermEditChallenge,
		PermDeleteChallenge,
		PermViewGame,
		PermJoinGame,
		PermSubmitFlag,
		PermViewOwnStats,
		PermViewRankings,
	},
	RoleOrganizer: {
		PermCreateGame,
		PermEditGame,
		PermDeleteGame,
		PermStartGame,
		PermPauseGame,
		PermStopGame,
		PermViewGameStats,
		PermManageTeams,
		PermCreateChallenge,
		PermEditChallenge,
		PermDeleteChallenge,
		PermViewGame,
		PermViewRankings,
	},
	RolePlayer: {
		PermViewGame,
		PermJoinGame,
		PermSubmitFlag,
		PermViewOwnStats,
		PermViewRankings,
	},
}

// GetRoleFromString converts string to Role type
func GetRoleFromString(roleStr string) Role {
	switch roleStr {
	case string(RoleAdmin):
		return RoleAdmin
	case string(RoleOrganizer):
		return RoleOrganizer
	default:
		return RolePlayer
	}
}
