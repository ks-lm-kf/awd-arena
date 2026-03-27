package database

import (
	"os"
	"path/filepath"

	"github.com/awd-platform/awd-arena/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB(dbPath string) error {
	if dbPath == "" {
		dbPath = "data/awd.db"
	}
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	return db.AutoMigrate(
		&model.Team{},
		&model.User{},
		&model.Game{},
		&model.GameTeam{},
		&model.FlagRecord{},
		&model.FlagSubmission{},
		&model.RoundScore{},
		&model.Challenge{},
		&model.TeamContainer{},
		&model.EventLog{},
		&model.DockerImage{},
		&model.AdminLog{},
	)
}

func GetDB() *gorm.DB {
	return db
}
