package model

import "time"
import "gorm.io/gorm"

type ChallengeTemplate struct {
    ID           int64             `json:"id" gorm:"primaryKey"`
    Name         string            `json:"name" gorm:"uniqueIndex;not null"`
    Category     string            `json:"category"`
    Description  string            `json:"description"`
    ImageConfig  ImageConfig       `json:"image_config" gorm:"type:json"`
    ServicePorts ServicePorts      `json:"service_ports" gorm:"type:json"`
    VulnConfig   VulnConfig        `json:"vuln_config" gorm:"type:json"`
    FlagConfig   FlagConfig        `json:"flag_config" gorm:"type:json"`
    Difficulty   string            `json:"difficulty" gorm:"default:medium"`
    BaseScore    int               `json:"base_score" gorm:"default:100"`
    CPULimit     float64           `json:"cpu_limit" gorm:"default:0.5"`
    MemLimit     int               `json:"mem_limit" gorm:"default:256"`
    Hints        string            `json:"hints"`
    Status       string            `json:"status" gorm:"default:draft"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
    DeletedAt    gorm.DeletedAt    `json:"-" gorm:"index"`
}

type ImageConfig struct {
    Name        string            `json:"name"`
    Tag         string            `json:"tag"`
    Repository  string            `json:"repository"`
    EnvVars     map[string]string `json:"env_vars"`
    Volumes     []VolumeMount     `json:"volumes"`
    NetworkMode string            `json:"network_mode"`
}

type VolumeMount struct {
    HostPath  string `json:"host_path"`
    ContainerPath string `json:"container_path"`
    ReadOnly  bool   `json:"read_only"`
}

type ServicePorts []ServicePort

type ServicePort struct {
    Port        int    `json:"port"`
    Protocol    string `json:"protocol"`
    ServiceName string `json:"service_name"`
    Description string `json:"description"`
}

type VulnConfig struct {
    Type        string   `json:"type"`
    CVE         string   `json:"cve"`
    CWE         string   `json:"cwe"`
    Severity    string   `json:"severity"`
    FixSuggestion string `json:"fix_suggestion"`
}

type FlagConfig struct {
    Type     string       `json:"type"`
    Value    string       `json:"value"`
    Format   string       `json:"format"`
    Location FlagLocation `json:"location"`
}

type FlagLocation struct {
    Type string `json:"type"`
    Path string `json:"path"`
}

type TemplateExport struct {
    Templates []ChallengeTemplate `json:"templates"`
    Version   string              `json:"version"`
    Exported  time.Time           `json:"exported"`
}

type TemplateImport struct {
    Templates []ChallengeTemplate `json:"templates"`
}

type TemplatePreview struct {
    Name        string `json:"name"`
    Category    string `json:"category"`
    Difficulty  string `json:"difficulty"`
    BaseScore   int    `json:"base_score"`
    Description string `json:"description"`
}
