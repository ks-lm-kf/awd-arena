package server

import (
	"github.com/awd-platform/awd-arena/internal/handler"
	"github.com/awd-platform/awd-arena/internal/middleware"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes sets up all HTTP routes.
func RegisterRoutes(app *fiber.App) {
	// API v1
	v1 := app.Group("/api/v1")

	// Auth (login has rate limiting to prevent brute force)
	auth := v1.Group("/auth")
	auth.Post("/login", middleware.LoginRateLimit(), handler.AuthHandler.Login)
	auth.Post("/logout", middleware.JWTAuth(), handler.AuthHandler.Logout)
	auth.Get("/me", middleware.JWTAuth(), handler.AuthHandler.Me)
	auth.Post("/register", handler.AuthHandler.Register)
	auth.Post("/refresh", middleware.JWTAuth(), handler.AuthHandler.RefreshToken)

	// Password change
	auth.Put("/change-password", middleware.JWTAuth(), handler.AuthHandler.ChangePassword)
	auth.Post("/password", middleware.JWTAuth(), handler.AuthHandler.ChangePassword)

	// Admin routes - User management
	admin := v1.Group("/admin", middleware.JWTAuth(), middleware.RequireRole(model.RoleAdmin))
	admin.Get("/users", handler.UserHandler.List)
	admin.Post("/users", handler.UserHandler.Create)
	admin.Get("/users/:id", handler.UserHandler.GetUser)
	admin.Put("/users/:id", handler.UserHandler.Update)
	admin.Delete("/users/:id", handler.UserHandler.Delete)

	// Security admin routes (view WAF rules, attack logs)
	security := v1.Group("/security", middleware.JWTAuth(), middleware.RequirePermission(model.PermViewAllData))
	security.Get("/waf/rules", handler.GetWAFRules)
	security.Get("/attacks", handler.GetGameAttacks)

	// Teams - All authenticated users can list teams
	teams := v1.Group("/teams", middleware.JWTAuth())
	teams.Get("/", middleware.RequirePermission(model.PermViewRankings), handler.TeamHandler.List)
	teams.Post("/", middleware.RequirePermission(model.PermManageTeams), handler.TeamHandler.Create)
	teams.Get("/:id", middleware.RequirePermission(model.PermViewRankings), handler.TeamHandler.Get)
	teams.Get("/:id/members", middleware.RequirePermission(model.PermViewRankings), handler.TeamHandler.Members)
	teams.Post("/:id/members", middleware.RequirePermission(model.PermManageTeams), handler.TeamHandler.AddMember)
	teams.Delete("/:id/members/:userId", middleware.RequirePermission(model.PermManageTeams), handler.TeamHandler.RemoveMember)
	teams.Delete("/:id", middleware.RequirePermission(model.PermManageTeams), handler.AdminHandler.DeleteTeam)

	// Games - Players can view, Organizers can manage
	games := v1.Group("/games", middleware.JWTAuth())
	games.Get("/", middleware.RequirePermission(model.PermViewGame), handler.GameHandler.List)
	games.Post("/", middleware.RequirePermission(model.PermCreateGame), handler.GameHandler.Create)
	games.Get("/:id", middleware.RequirePermission(model.PermViewGame), handler.GameHandler.Get)
	games.Put("/:id", middleware.RequirePermission(model.PermEditGame), handler.GameHandler.Update)
	games.Delete("/:id", middleware.RequirePermission(model.PermManageGames), handler.GameHandler.Delete)
	games.Post("/:id/start", middleware.RequirePermission(model.PermStartGame), handler.GameHandler.Start)
	games.Post("/:id/pause", middleware.RequirePermission(model.PermPauseGame), handler.GameHandler.Pause)
	games.Post("/:id/resume", middleware.RequirePermission(model.PermPauseGame), handler.GameHandler.Resume)
	games.Post("/:id/stop", middleware.RequirePermission(model.PermStopGame), handler.GameHandler.Stop)
	games.Post("/:id/reset", middleware.RequirePermission(model.PermManageGames), handler.GameHandler.Reset)

	// Challenges - Players can view, Organizers can manage
	games.Get("/:id/challenges", middleware.RequirePermission(model.PermViewGame), handler.ChallengeHandler.List)
	games.Post("/:id/challenges", middleware.RequirePermission(model.PermCreateChallenge), handler.ChallengeHandler.Create)
	games.Put("/:id/challenges/:challengeId", middleware.RequirePermission(model.PermEditGame), handler.ChallengeHandler.Update)
	games.Delete("/:id/challenges/:challengeId", middleware.RequirePermission(model.PermManageGames), handler.ChallengeHandler.Delete)

	// Security: alerts & attacks per game
	games.Get("/:id/alerts", middleware.RequirePermission(model.PermViewGameStats), handler.GetGameAlerts)
	games.Get("/:id/attacks", middleware.RequirePermission(model.PermViewGameStats), handler.GetGameAttacks)

	// Containers - Admin and Organizers can view/restart
	games.Get("/:id/containers", middleware.RequirePermission(model.PermViewGameStats), handler.ContainerHandler.List)
	games.Post("/:id/containers/restart", middleware.RequirePermission(model.PermManageInfrastructure), handler.ContainerHandler.BulkRestart)
	games.Post("/:id/containers/:cid/restart", middleware.RequirePermission(model.PermManageInfrastructure), handler.ContainerHandler.Restart)
	games.Get("/:id/containers/stats", middleware.RequirePermission(model.PermViewGameStats), handler.ContainerHandler.Stats)

	// Flags - Players can submit flags
	flags := v1.Group("/games/:id/flags", middleware.JWTAuth())
	flags.Post("/submit", middleware.FlagSubmitRateLimit(100, 60), middleware.RequirePermission(model.PermSubmitFlag), handler.FlagHandler.Submit)
	flags.Get("/history", middleware.RequirePermission(model.PermViewOwnStats), handler.FlagHandler.History)

	// Rankings - All authenticated users can view
	v1.Get("/games/:id/rankings", middleware.JWTAuth(), middleware.RequirePermission(model.PermViewRankings), handler.RankingHandler.Get)
	v1.Get("/games/:id/rankings/rounds/:round", middleware.JWTAuth(), middleware.RequirePermission(model.PermViewRankings), handler.RankingHandler.GetRound)

	// Player container info - Players can view their own containers
	games.Get("/:id/my-containers", middleware.RequirePermission(model.PermViewOwnStats), handler.ScoreHandler.GetMyContainers)
	games.Get("/:id/my-machines", middleware.RequirePermission(model.PermViewOwnStats), handler.ScoreHandler.GetMyContainers)

	// WebSocket (token validated via query param in ws.go)
	app.Get("/ws", HandleWebSocket)

	// Docker Images - Legacy routes (keeping for compatibility)
	dockerImages := v1.Group("/docker-images", middleware.JWTAuth(), middleware.RequireRole(model.RoleAdmin))
	dockerImages.Get("/", handler.DockerImageHandlerObj.List)
	dockerImages.Get("/host/list", handler.DockerImageHandlerObj.HostList)
	dockerImages.Get("/:id", handler.DockerImageHandlerObj.Get)
	dockerImages.Post("/", handler.DockerImageHandlerObj.Create)
	dockerImages.Put("/:id", handler.DockerImageHandlerObj.Update)
	dockerImages.Delete("/:id", handler.DockerImageHandlerObj.Delete)
	dockerImages.Post("/:id/pull", handler.DockerImageHandlerObj.Pull)

	// Admin Image Management Routes - Enhanced with full CRUD operations
	adminImages := v1.Group("/admin/images", middleware.JWTAuth(), middleware.RequireRole(model.RoleAdmin))
	adminImages.Get("/", handler.DockerImageHandlerObj.List)
	adminImages.Get("/host/list", handler.DockerImageHandlerObj.HostList)
	adminImages.Post("/pull", handler.DockerImageHandlerObj.PullImage)
	adminImages.Post("/push", handler.DockerImageHandlerObj.PushImage)
	adminImages.Post("/build", handler.DockerImageHandlerObj.BuildImage)
	adminImages.Get("/:id", handler.DockerImageHandlerObj.Get)
	adminImages.Get("/:id/details", handler.DockerImageHandlerObj.GetImageDetails)
	adminImages.Delete("/:id", handler.DockerImageHandlerObj.Delete)
	adminImages.Delete("/:id/complete", handler.DockerImageHandlerObj.RemoveFromDBAndHost)
	adminImages.Delete("/host/:id", handler.DockerImageHandlerObj.RemoveFromHost)
	adminImages.Post("/:id/pull", handler.DockerImageHandlerObj.Pull)
	adminImages.Post("/", handler.DockerImageHandlerObj.Create)
	adminImages.Put("/:id", handler.DockerImageHandlerObj.Update)

	// Rounds - Round management
	games.Get("/:id/rounds", middleware.RequirePermission(model.PermViewGame), handler.RoundHandler.GetRounds)
	games.Post("/:id/rounds", middleware.RequirePermission(model.PermPauseGame), handler.RoundHandler.ControlRounds)

	// Admin routes for judges (organizer role) - Enhanced with logging
	judge := v1.Group("/judge", middleware.JWTAuth(), middleware.RequireRole(model.RoleAdmin, model.RoleOrganizer))

	// Admin logs
	judge.Get("/logs", handler.AdminHandler.GetAdminLogs)

	// Game management with logging
	judge.Post("/games", handler.AdminHandler.CreateGame)
	judge.Put("/games/:id", handler.AdminHandler.UpdateGame)
	judge.Delete("/games/:id", handler.AdminHandler.DeleteGame)
	judge.Post("/games/:id/start", handler.AdminHandler.StartGame)
	judge.Post("/games/:id/pause", handler.AdminHandler.PauseGame)
	judge.Post("/games/:id/resume", handler.AdminHandler.ResumeGame)
	judge.Post("/games/:id/stop", handler.AdminHandler.StopGame)
	judge.Post("/games/:id/reset", handler.AdminHandler.ResetGame)
	judge.Post("/games/:id/teams", handler.AdminHandler.AddTeamToGame)
	judge.Get("/games/:id/teams", handler.AdminHandler.GetGameTeams)
	judge.Delete("/games/:id/teams/:team_id", handler.AdminHandler.RemoveTeamFromGame)

	// Team management with logging
	judge.Post("/teams", handler.AdminHandler.CreateTeam)
	judge.Put("/teams/:id", handler.AdminHandler.UpdateTeam)
	judge.Delete("/teams/:id", handler.AdminHandler.DeleteTeam)
	judge.Post("/teams/batch-import", handler.AdminHandler.BatchImportTeams)
	judge.Post("/teams/:id/reset", handler.AdminHandler.ResetTeam)

	// Score adjustment
	judge.Post("/scores/adjust", handler.AdminHandler.AdjustScore)

	// Settings - System settings management
	settings := v1.Group("/settings", middleware.JWTAuth())
	settings.Get("/", handler.SettingsHandler.GetSettings)
	settings.Put("/", middleware.RequirePermission(model.PermManageSettings), handler.SettingsHandler.UpdateSettings)

	// WAF rules alias (test expects /api/v1/waf/rules)
	// WAF rules - use group-level auth
	wafGroup := v1.Group("/waf", middleware.JWTAuth(), middleware.RequireRole(model.RoleAdmin))
	wafGroup.Get("/rules", handler.GetWAFRules)

	// Health
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Dashboard routes
	dashboard := v1.Group("/dashboard", middleware.JWTAuth())
	dashboard.Get("/", handler.DashboardHandler.GetDashboard)
	dashboard.Get("/activity", handler.DashboardHandler.GetRecentActivity)

	// Audit API (admin only)
	audit := app.Group("/api/audit", middleware.JWTAuth(), middleware.RequireRole(model.RoleAdmin))
	audit.Get("/logs", handler.AuditHandler.GetAuditLogs)
	audit.Get("/stats", handler.AuditHandler.GetAuditStats)

	// Export API
	export := v1.Group("/games/:id/export", middleware.JWTAuth())
	export.Get("/scoreboard/csv", handler.ExportHandler.ExportRankingCSV)
	export.Get("/ranking/csv", handler.ExportHandler.ExportRankingCSV)
	export.Get("/scoreboard/pdf", handler.ExportHandler.ExportRankingPDF)
	export.Get("/ranking/pdf", handler.ExportHandler.ExportRankingPDF)
	export.Get("/attacks", handler.ExportHandler.ExportAttackLog)
	export.Get("/all", handler.ExportHandler.ExportAll)
}
