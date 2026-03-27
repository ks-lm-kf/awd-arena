package repo

import (
	"context"

	"github.com/awd-platform/awd-arena/internal/model"
)

// GameTeam represents the game-team junction.
type GameTeam struct {
	GameID int64   `json:"game_id"`
	TeamID int64   `json:"team_id"`
	Score  float64 `json:"score"`
	Rank   int     `json:"rank"`
}

// GameRepo defines game data access operations.
type GameRepo interface {
	Create(ctx context.Context, game *model.Game) error
	GetByID(ctx context.Context, id int64) (*model.Game, error)
	List(ctx context.Context, offset, limit int) ([]*model.Game, error)
	Update(ctx context.Context, game *model.Game) error
	Delete(ctx context.Context, id int64) error
	AddTeam(ctx context.Context, gameID, teamID int64) error
	GetTeams(ctx context.Context, gameID int64) ([]*GameTeam, error)
}
