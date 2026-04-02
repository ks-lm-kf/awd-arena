package engine

import (
	"context"
	"fmt"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/service"
	"github.com/awd-platform/awd-arena/pkg/crypto"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// FlagManager manages flag generation and validation.
type FlagManager struct {
	game    *model.Game
	flagSvc *service.FlagService
	flags   map[string]*model.FlagRecord // key: "round:teamID:service"
}

func NewFlagManager(game *model.Game) *FlagManager {
	return &FlagManager{
		game:    game,
		flagSvc: &service.FlagService{},
		flags:   make(map[string]*model.FlagRecord),
	}
}

func (fm *FlagManager) GenerateRoundFlags(ctx context.Context, round int) error {
	logger.Info("generating round flags", "round", round)
	_, err := fm.flagSvc.GenerateFlags(ctx, fm.game.ID, round)
	return err
}

func (fm *FlagManager) GenerateFlag(teamID int64, service string, round int) string {
	format := fm.game.FlagFormat
	if format == "" {
		format = "flag{%s}"
	}
	return crypto.GenerateFlag(format, fmt.Sprintf("%d", teamID), service, round)
}

func (fm *FlagManager) ValidateFlag(flag string) (*model.FlagRecord, bool) {
	flagHash := crypto.SHA256Hex(flag)
	for _, record := range fm.flags {
		if record.FlagHash == flagHash {
			return record, true
		}
	}
	db := database.GetDB()
	if db != nil {
		var record model.FlagRecord
		if err := db.Where("flag_hash = ?", flagHash).First(&record).Error; err == nil {
			return &record, true
		}
	}
	return nil, false
}

func (fm *FlagManager) GetCurrentFlags() []*model.FlagRecord {
	result := make([]*model.FlagRecord, 0, len(fm.flags))
	for _, f := range fm.flags {
		result = append(result, f)
	}
	return result
}
