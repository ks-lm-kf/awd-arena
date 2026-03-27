package repo

import (
	"context"

	"github.com/awd-platform/awd-arena/internal/model"
)

// FlagRepo defines flag data access operations.
type FlagRepo interface {
	SaveFlag(ctx context.Context, record *model.FlagRecord) error
	GetFlag(ctx context.Context, gameID int64, round int, teamID int64, service string) (*model.FlagRecord, error)
	SubmitFlag(ctx context.Context, sub *model.FlagSubmission) error
	GetSubmissions(ctx context.Context, gameID int64, round int) ([]*model.FlagSubmission, error)
}
