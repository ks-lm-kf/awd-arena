package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	NATS     NATSConfig     `yaml:"nats"`
	Docker   DockerConfig   `yaml:"docker"`
	Security SecurityConfig `yaml:"security"`
	Game     GameConfig     `yaml:"game"`
	AI       AIConfig       `yaml:"ai"`
}

type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	TLSPort   int    `yaml:"tls_port"`
	JWTSecret string `yaml:"jwt_secret"`
	StaticDir string `yaml:"static_dir"`
}

type DatabaseConfig struct {
	SQLitePath string `yaml:"sqlite_path"`
}

type NATSConfig struct {
	URL     string `yaml:"url"`
	Cluster string `yaml:"cluster"`
}

type DockerConfig struct {
	Host string `yaml:"host"`
}

type SecurityConfig struct {
	JWTExpireHours   int  `yaml:"jwt_expire_hours"`
	RateLimit        int  `yaml:"rate_limit"`
	FlagSubmitLimit  int  `yaml:"flag_submit_limit"`
	BlocklistEnabled bool `yaml:"blocklist_enabled"`
}

type GameConfig struct {
	DefaultRoundDuration time.Duration `yaml:"default_round_duration"`
	DefaultBreakDuration time.Duration `yaml:"default_break_duration"`
	MaxContainersPerTeam int           `yaml:"max_containers_per_team"`
}

type AIConfig struct {
	Enabled        bool   `yaml:"enabled"`
	RuleEnginePath string `yaml:"rule_engine_path"`
	ONNXModelPath  string `yaml:"onnx_model_path"`
}

var C Config

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &C); err != nil {
		return nil, err
	}
	setDefaults()

	if C.Server.JWTSecret == "" || len(C.Server.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	defaultSecrets := map[string]bool{
		"CHANGE_ME_TO_A_STRONG_RANDOM_SECRET": true,
		"dev-secret-change-me":                true,
		"your-strong-random-secret-here":      true,
		"secret":                              true,
	}
	if defaultSecrets[C.Server.JWTSecret] {
		return nil, fmt.Errorf("ERROR: JWT secret is still the default value (%q). Set a strong random secret in config.yaml or JWT_SECRET env var", C.Server.JWTSecret)
	}

	return &C, nil
}

func setDefaults() {
	if C.Server.Host == "" {
		C.Server.Host = "0.0.0.0"
	}
	if C.Server.Port == 0 {
		C.Server.Port = 8080
	}
	if C.Database.SQLitePath == "" {
		C.Database.SQLitePath = "data/awd.db"
	}
	if C.Server.StaticDir == "" {
		C.Server.StaticDir = "web/dist"
	}
	if C.Security.JWTExpireHours == 0 {
		C.Security.JWTExpireHours = 2
	}
	if C.Security.RateLimit == 0 {
		C.Security.RateLimit = 100
	}
	if C.Game.DefaultRoundDuration == 0 {
		C.Game.DefaultRoundDuration = 5 * time.Minute
	}
	if C.Game.DefaultBreakDuration == 0 {
		C.Game.DefaultBreakDuration = 2 * time.Minute
	}
}
