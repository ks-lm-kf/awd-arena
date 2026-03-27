package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/awd-platform/awd-arena/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	sqlFile := flag.String("sql", "migrations/001_init_schema.sql", "path to SQL file")
	flag.Parse()

	logger.Init("info")
	logger.Info("running database migration", "config", *configPath, "sql", *sqlFile)

	f, err := os.Open(*sqlFile)
	if err != nil {
		logger.Error("failed to open SQL file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	sql, err := io.ReadAll(f)
	if err != nil {
		logger.Error("failed to read SQL file", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Migration SQL loaded: %d bytes\n", len(sql))
	fmt.Println("Connect to database and execute migration.")
	fmt.Println("(Database connection will be implemented with pgx)")
}
