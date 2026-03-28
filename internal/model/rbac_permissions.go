package model

type Role string

const (
    RoleAdmin     Role = "admin"
    RoleOrganizer Role = "organizer"
    RolePlayer    Role = "player"
)

type Permission string

const (
    ManageUsers          Permission = "manage_users"
    ManageGames          Permission = "manage_games"
    CreateGame           Permission = "create_game"
    EditGame             Permission = "edit_game"
    DeleteGame           Permission = "delete_game"
    StartGame            Permission = "start_game"
    PauseGame            Permission = "pause_game"
    StopGame             Permission = "stop_game"
    ViewGame             Permission = "view_game"
    ManageChallenges     Permission = "manage_challenges"
    CreateChallenge      Permission = "create_challenge"
    EditChallenge        Permission = "edit_challenge"
    DeleteChallenge      Permission = "delete_challenge"
    ManageTeams          Permission = "manage_teams"
    SubmitFlag           Permission = "submit_flag"
    ViewGameStats        Permission = "view_game_stats"
    ViewOwnStats         Permission = "view_own_stats"
    ViewRankings         Permission = "view_rankings"
    ManageInfrastructure Permission = "manage_infrastructure"
    ManageSettings       Permission = "manage_settings"
    ViewAllData          Permission = "view_all_data"
)


// Perm aliases for handler convenience
var (
    PermManageUsers          = ManageUsers
    PermManageGames          = ManageGames
    PermCreateGame           = CreateGame
    PermEditGame             = EditGame
    PermDeleteGame           = DeleteGame
    PermStartGame            = StartGame
    PermPauseGame            = PauseGame
    PermStopGame             = StopGame
    PermViewGame             = ViewGame
    PermManageChallenges     = ManageChallenges
    PermCreateChallenge      = CreateChallenge
    PermEditChallenge        = EditChallenge
    PermDeleteChallenge      = DeleteChallenge
    PermManageTeams          = ManageTeams
    PermSubmitFlag           = SubmitFlag
    PermViewGameStats        = ViewGameStats
    PermViewOwnStats         = ViewOwnStats
    PermViewRankings         = ViewRankings
    PermManageInfrastructure = ManageInfrastructure
    PermManageSettings       = ManageSettings
    PermViewAllData          = ViewAllData
)

var RolePermissions = map[Role][]Permission{
    RoleAdmin: {
        ManageUsers, ManageGames, CreateGame, EditGame, DeleteGame,
        StartGame, PauseGame, StopGame, ViewGame,
        ManageChallenges, CreateChallenge, EditChallenge, DeleteChallenge,
        ManageTeams, SubmitFlag, ViewGameStats, ViewOwnStats,
        ViewRankings, ManageInfrastructure, ManageSettings, ViewAllData,
    },
    RoleOrganizer: {
        CreateGame, EditGame, StartGame, PauseGame, StopGame, ViewGame,
        ManageChallenges, CreateChallenge, EditChallenge, DeleteChallenge,
        ManageTeams, ViewGameStats, ViewRankings,
    },
    RolePlayer: {
        SubmitFlag, ViewOwnStats, ViewRankings, ViewGame,
    },
}
