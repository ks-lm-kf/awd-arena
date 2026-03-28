package server

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/awd-platform/awd-arena/internal/config"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/security"
	"github.com/awd-platform/awd-arena/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// Server wraps the HTTP server.
type Server struct {
	app *fiber.App
	cfg *config.Config
}

// New creates a new Server instance.
func New(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{})
	app.Use(recover.New())
	
	// Security headers - apply first
	app.Use(middleware.SecurityHeaders())
	
	app.Use(middleware.CORS())
	app.Use(middleware.Logger())

	// Initialize WAF and add as global middleware
	middleware.InitWAF(security.NewAttackLogStore(1000))
	app.Use(middleware.WAFMiddleware())

	// Global API rate limit (100 req/min per IP)
	app.Use("/api/", middleware.GlobalAPIRateLimit(100, 1*time.Minute))

	srv := &Server{app: app, cfg: cfg}
	RegisterRoutes(app)

	// Register WebSocket hub as EventBus broadcaster
	eventbus.SetBroadcaster(Hub)

	// Serve frontend static files
	staticDir := cfg.Server.StaticDir
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		app.Use(static.New(staticDir, static.Config{
			IndexNames: []string{"index.html"},
		}))

		// SPA fallback: serve index.html for all non-API, non-WebSocket routes
		app.Get("/*", func(c fiber.Ctx) error {
			// Skip API routes and WebSocket route
			path := c.Path()
			if len(path) >= 4 && path[:4] == "/api" {
				return c.SendStatus(404) // API route not found
			}
			if path == "/ws" {
				return c.SendStatus(404) // WebSocket route not found
			}
			
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				return c.SendFile(indexPath)
			}
			return c.SendStatus(404)
		})
	}

	return srv
}

// Start begins listening.
func (s *Server) Start() error {
	addr := s.cfg.Server.Host + ":" + strconv.Itoa(s.cfg.Server.Port)
	return s.app.Listen(addr)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}

// App returns the Fiber app.
func (s *Server) App() *fiber.App {
	return s.app
}
