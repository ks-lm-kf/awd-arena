package mode

import (
	"context"

	"github.com/awd-platform/awd-arena/internal/model"
)

// GameMode defines the interface for competition modes.
type GameMode interface {
	Start(ctx context.Context, game *model.Game) error
	OnRoundStart(ctx context.Context, round int) error
	OnRoundEnd(ctx context.Context, round int) error
	OnAttack(ctx context.Context, attack *model.FlagSubmission) error
	OnDefense(ctx context.Context, teamID int64, flag string) error
	CalculateScore(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ModeFactory creates a GameMode by name.
func ModeFactory(name string) GameMode {
	switch name {
	case "awd_score":
		return NewAWDScoreMode()
	case "awd_mix":
		return NewAWDMixMode()
	case "koh":
		return NewKingOfHillMode()
	default:
		return NewAWDScoreMode()
	}
}
