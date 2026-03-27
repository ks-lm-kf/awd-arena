package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ChallengeTemplate 题库模板
type ChallengeTemplate struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:200;not null;uniqueIndex"`
	Category    string    `json:"category" gorm:"size:50;not null;index"` // web, pwn, crypto, misc, reverse
	Description string    `json:"description" gorm:"type:text"`
	
	// Docker镜像配置
	ImageConfig ImageConfig `json:"image_config" gorm:"type:jsonb"`
	
	// 服务端口配置
	ServicePorts ServicePorts `json:"service_ports" gorm:"type:jsonb"`
	
	// 漏洞配置
	VulnConfig VulnConfig `json:"vuln_config" gorm:"type:jsonb"`
	
	// Flag配置
	FlagConfig FlagConfig `json:"flag_config" gorm:"type:jsonb"`
	
	// 难度和分数
	Difficulty  string `json:"difficulty" gorm:"size:20;default:medium"` // easy, medium, hard
	BaseScore   int    `json:"base_score" gorm:"default:100"`
	
	// 资源限制
	CPULimit   float64 `json:"cpu_limit" gorm:"default:0.5"`
	MemLimit   int     `json:"mem_limit" gorm:"default:256"` // MB
	
	// 提示信息
	Hints string `json:"hints" gorm:"type:text"`
	
	// 状态
	Status    string    `json:"status" gorm:"size:20;default:draft"` // draft, published, archived
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ImageConfig Docker镜像配置
type ImageConfig struct {
	ImageName   string            `json:"image_name"`              // 镜像名称
	ImageTag    string            `json:"image_tag"`               // 镜像标签
	Registry    string            `json:"registry"`                // 镜像仓库地址
	Environment map[string]string `json:"environment"`             // 环境变量
	Volumes     []VolumeMount     `json:"volumes"`                 // 挂载卷
	NetworkMode string            `json:"network_mode"`            // 网络模式
	Privileged  bool              `json:"privileged"`              // 是否特权模式
}

// VolumeMount 卷挂载配置
type VolumeMount struct {
	HostPath    string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	ReadOnly    bool   `json:"read_only"`
}

// ServicePorts 服务端口配置
type ServicePorts []ServicePort

// ServicePort 单个服务端口
type ServicePort struct {
	Port        int    `json:"port"`          // 容器内端口
	Protocol    string `json:"protocol"`      // tcp, udp
	ServiceName string `json:"service_name"`  // 服务名称
	Description string `json:"description"`   // 服务描述
	IsPrimary   bool   `json:"is_primary"`    // 是否主要服务
}

// VulnConfig 漏洞配置
type VulnConfig struct {
	VulnType      string            `json:"vuln_type"`       // 漏洞类型
	CVE           string            `json:"cve"`             // CVE编号
	CWE           string            `json:"cwe"`             // CWE分类
	Severity      string            `json:"severity"`        // 严重程度
	AttackVector  string            `json:"attack_vector"`   // 攻击向量
	Solution      string            `json:"solution"`        // 修复方案
	References    []string          `json:"references"`      // 参考链接
	Tags          []string          `json:"tags"`            // 标签
	ExploitHints  string            `json:"exploit_hints"`   // 利用提示
}

// FlagConfig Flag配置
type FlagConfig struct {
	FlagType    string      `json:"flag_type"`     // static, dynamic, regex
	FlagValue   string      `json:"flag_value"`    // 静态flag值或生成规则
	FlagFormat  string      `json:"flag_format"`   // flag格式，如 flag{...}
	Location    FlagLocation `json:"location"`     // flag位置
	Points      int         `json:"points"`        // 分值
}

// FlagLocation Flag位置配置
type FlagLocation struct {
	Type        string `json:"type"`         // file, env, database, api
	Path        string `json:"path"`         // 文件路径或环境变量名
	Permissions string `json:"permissions"`  // 权限设置
	Command     string `json:"command"`      // 获取命令（如果需要）
}

// 实现driver.Valuer接口，用于GORM JSON存储
func (ic ImageConfig) Value() (driver.Value, error) {
	return json.Marshal(ic)
}

func (ic *ImageConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, ic)
}

func (sp ServicePorts) Value() (driver.Value, error) {
	return json.Marshal(sp)
}

func (sp *ServicePorts) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, sp)
}

func (vc VulnConfig) Value() (driver.Value, error) {
	return json.Marshal(vc)
}

func (vc *VulnConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, vc)
}

func (fc FlagConfig) Value() (driver.Value, error) {
	return json.Marshal(fc)
}

func (fc *FlagConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, fc)
}

// TemplateExport 导出模板的JSON格式
type TemplateExport struct {
	Version     string             `json:"version"`
	ExportedAt  time.Time          `json:"exported_at"`
	Template    ChallengeTemplate  `json:"template"`
}

// TemplateImport 导入模板请求
type TemplateImport struct {
	Template TemplateExport `json:"template"`
	Overwrite bool          `json:"overwrite"` // 是否覆盖已存在的同名模板
}

// TemplateListQuery 模板列表查询参数
type TemplateListQuery struct {
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
	Status     string `form:"status"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

// TemplatePreview 模板预览
type TemplatePreview struct {
	ChallengeTemplate
	DockerCommand string            `json:"docker_command"` // 生成的Docker命令
	PortMapping   map[int]int       `json:"port_mapping"`   // 端口映射预览
	EnvList       []string          `json:"env_list"`       // 环境变量列表
}

