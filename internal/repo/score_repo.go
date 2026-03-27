package repo

import (
	"context"

	"github.com/awd-platform/awd-arena/internal/model"
)

// ScoreRepo defines score data access operations.
type ScoreRepo interface {
	SaveRoundScore(ctx context.Context, score *model.RoundScore) error
	GetRoundScores(ctx context.Context, gameID int64, round int) ([]*model.RoundScore, error)
	GetTeamScores(ctx context.Context, gameID int64) ([]*model.RoundScore, error)
}
