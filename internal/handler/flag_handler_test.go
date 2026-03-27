package handler

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestFlagRefreshHandler_RefreshFlags_InvalidGameID(t *testing.T) {
	app := fiber.New()
	handler := NewFlagRefreshHandler(nil)

	app.Post("/games/:id/flags/refresh", handler.RefreshFlags)

	req := httptest.NewRequest("POST", "/games/invalid/flags/refresh", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestFlagRefreshHandler_RefreshFlags_InvalidBody(t *testing.T) {
	app := fiber.New()
	handler := NewFlagRefreshHandler(nil)

	app.Post("/games/:id/flags/refresh", handler.RefreshFlags)

	req := httptest.NewRequest("POST", "/games/1/flags/refresh", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestFlagRefreshHandler_RefreshFlags_MissingFields(t *testing.T) {
	// Skip this test as it requires database initialization
	// In production, this should use a mock database
	t.Skip("requires database initialization")
}

func TestFlagRefreshHandler_GetFlags_InvalidGameID(t *testing.T) {
	// Skip this test as it requires database initialization
	t.Skip("requires database initialization")
}

func TestFlagRefreshHandler_NewFlagRefreshHandler(t *testing.T) {
	handler := NewFlagRefreshHandler(nil)
	if handler == nil {
		t.Fatal("NewFlagRefreshHandler returned nil")
	}
	if handler.generator == nil {
		t.Error("generator is nil")
	}
	if handler.writer == nil {
		t.Error("writer is nil")
	}
}
