# AWD 靶场平台 — 技术架构设计文档

> **版本**: v1.0  
> **日期**: 2026-03-19  
> **作者**: Architect Agent  
> **状态**: Draft

---

## 目录

1. [整体架构](#1-整体架构)
2. [技术栈选型](#2-技术栈选型)
3. [核心模块设计](#3-核心模块设计)
4. [跨平台方案](#4-跨平台方案)
5. [数据模型](#5-数据模型)
6. [API 设计](#6-api-设计)
7. [部署架构](#7-部署架构)
8. [项目目录结构](#8-项目目录结构)

---

## 1. 整体架构

### 1.1 系统架构图（文字描述）

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端 (前端)                            │
│  React + TypeScript  │  实时大屏  │  管理后台  │  WebSocket 客户端  │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS / WSS
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API 网关层 (Go)                             │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐  ┌──────────────┐   │
│  │ JWT Auth │  │ Rate     │  │ WAF       │  │ Reverse      │   │
│  │ 中间件    │  │ Limiter  │  │ 基础检测   │  │ Proxy       │   │
│  └──────────┘  └──────────┘  └───────────┘  └──────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │ gRPC / HTTP
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       业务逻辑层 (Go Microservices)              │
│                                                                 │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐  │
│  │ 竞赛引擎    │ │ 容器管理    │ │ 网络管理    │ │ 安全检测    │  │
│  │ Competition │ │ Container  │ │ Network    │ │ Security   │  │
│  │ Engine      │ │ Manager    │ │ Manager    │ │ Guard      │  │
│  └─────┬──────┘ └─────┬──────┘ └─────┬──────┘ └─────┬──────┘  │
│        │              │              │              │          │
│  ┌─────┴──────┐ ┌─────┴──────┐ ┌─────┴──────┐ ┌─────┴──────┐  │
│  │ 积分排名    │ │ 资源监控    │ │ 流量镜像    │ │ AI 分析     │  │
│  │ Scoring    │ │ Monitor    │ │ Mirror     │ │ Analyzer   │  │
│  └────────────┘ └────────────┘ └────────────┘ └────────────┘  │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              事件总线 (Event Bus / NATS)                   │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
┌──────────────────┐ ┌────────────┐ ┌──────────────────┐
│    数据层         │ │ 消息队列    │ │    基础设施层     │
│  ┌────────────┐  │ │            │ │                  │
│  │ PostgreSQL │  │ │  NATS Jet  │ │  Docker Engine   │
│  │ (关系数据)  │  │ │  Stream    │ │  (容器运行时)     │
│  ├────────────┤  │ │            │ │                  │
│  │ Redis      │  │ ├────────────┤ │  ┌────────────┐ │
│  │ (缓存/排名) │  │ │  Kafka    │ │  │ OVS/VLAN   │ │
│  ├────────────┤  │ │  (日志流)  │ │  │ 网络隔离    │ │
│  │ ClickHouse │  │ │            │ │  └────────────┘ │
│  │ (攻击日志)  │  │ └────────────┘ │                  │
│  ├────────────┤  │                │  ┌────────────┐ │
│  │ MinIO/S3   │  │                │  │ Prometheus │ │
│  │ (PCAP存储)  │  │                │  │ + Grafana  │ │
│  └────────────┘  │                │  └────────────┘ │
└──────────────────┘                └──────────────────┘
```

### 1.2 分层设计

| 层级 | 职责 | 技术 |
|------|------|------|
| **前端层** | 用户界面、实时大屏、管理后台 | React 19 + TypeScript 5 + Vite |
| **API 网关层** | 认证鉴权、限流、WAF、路由转发 | Go (Fiber) + Casbin |
| **业务层** | 竞赛逻辑、容器管理、安全检测、AI分析 | Go Microservices + gRPC |
| **数据层** | 持久化、缓存、时序数据、对象存储 | PostgreSQL + Redis + ClickHouse + MinIO |
| **基础设施层** | 容器运行时、网络隔离、监控告警 | Docker Engine + OVS + Prometheus |

---

## 2. 技术栈选型

### 2.1 方案对比

| 维度 | 🏆 方案A（推荐） | 方案B | 方案C |
|------|---------|-------|-------|
| **后端** | **Go 1.24+** | Rust | Python (FastAPI) |
| **前端** | React 19 + TypeScript | Vue 3 + TypeScript | React 19 + TypeScript |
| **Web框架** | **Fiber v3** | Actix-web | FastAPI |
| **ORM** | **GORM** + sqlc | SeaORM | SQLAlchemy |
| **数据库** | **PostgreSQL 17** + Redis 8 + ClickHouse | PostgreSQL 17 + Redis 8 + ClickHouse | PostgreSQL 17 + Redis 8 |
| **消息队列** | **NATS JetStream** (内置) + Kafka (可选) | NTex channels + Kafka | Celery + Redis |
| **容器管理** | **Docker SDK for Go** | bollard (Docker) | docker-py |
| **监控** | Prometheus + Grafana | Prometheus + Grafana | Prometheus + Grafana |
| **AI** | ONNX Runtime (本地推理) | candle / burn | PyTorch / scikit-learn |

### 2.2 推荐方案 A 详细说明

#### 后端：Go 1.24+
- **单二进制部署**，跨平台零依赖
- 原生并发 (goroutine)，轻松应对 50+ 队伍并发
- Docker SDK 官方支持最好
- 编译速度快，开发效率高
- 丰富的网络库，适合网络管理模块

#### 前端：React 19 + TypeScript 5 + Vite 6
- React 生态最大，组件库丰富
- TypeScript 保证类型安全
- Vite 极速开发体验
- 实时大屏用 Recharts + WebSocket

#### 数据库选型理由
| 数据库 | 用途 | 选型理由 |
|--------|------|----------|
| PostgreSQL 17 | 核心业务数据 | ACID、JSON支持、成熟稳定 |
| Redis 8 | 排行榜、缓存、实时排名 | ZSET 天然适配排名、Pub/Sub |
| ClickHouse | 攻击日志、流量分析 | 列式存储，亿级数据秒查 |
| MinIO | PCAP 文件、比赛快照 | S3 兼容，自托管 |

#### 消息队列：NATS JetStream
- 轻量级，单二进制，跨平台
- JetStream 提供持久化和 At-least-once 语义
- 内置 Pub/Sub + Queue Groups
- 延迟低 (< 1ms)，适合竞赛状态实时同步
- 后期流量大时可替换/补充 Kafka

---

## 3. 核心模块设计

### 3.1 引擎层 — 竞赛引擎

```
┌─────────────────────────────────────────────┐
│              CompetitionEngine               │
│                                             │
│  ┌─────────────┐    ┌──────────────────┐    │
│  │ RoundTimer  │───▶│ PhaseController  │    │
│  │ (定时调度)   │    │ (阶段管理)        │    │
│  └─────────────┘    └────────┬─────────┘    │
│                              │              │
│  ┌─────────────┐    ┌───────▼─────────┐    │
│  │ FlagManager │    │ ScoreCalculator │    │
│  │ (Flag生成)   │◀──▶│ (积分计算)       │    │
│  └─────────────┘    └───────┬─────────┘    │
│                              │              │
│  ┌─────────────┐    ┌───────▼─────────┐    │
│  │ AttackJudge │    │ EventPublisher  │    │
│  │ (攻击判定)   │───▶│ (事件发布)       │    │
│  └─────────────┘    └──────────────────┘    │
│                                             │
└─────────────────────────────────────────────┘
```

**Round 调度流程：**

```go
type RoundPhase int
const (
    PhasePreparation  RoundPhase = iota // 准备阶段
    PhaseRunning                       // 比赛进行中
    PhaseScoring                       // 结算阶段
    PhaseBreak                         // 休息阶段
)

type CompetitionEngine struct {
    currentRound  int
    currentPhase  RoundPhase
    roundDuration time.Duration       // 单轮时长 (默认 5min)
    breakDuration time.Duration       // 休息时长 (默认 2min)
    totalRounds   int                 // 总轮数
    ticker        *time.Ticker
    state         *GameState
}

// 竞赛模式可扩展接口
type GameMode interface {
    Start(ctx context.Context, game *Game) error
    OnRoundStart(ctx context.Context, round int) error
    OnRoundEnd(ctx context.Context, round int) error
    OnAttack(ctx context.Context, attack *AttackRecord) error
    OnDefense(ctx context.Context, defense *DefenseRecord) error
    CalculateScore(ctx context.Context) error
    Stop(ctx context.Context) error
}

// 已实现模式
// - AWDScoreMode    经典AWD攻防积分
// - AWDMixMode      攻防混合+解题
// - KingOfHillMode  山顶争夺
```

**Flag 管理策略：**
- 每轮自动轮换 Flag（SHA256 随机生成）
- Flag 存储在服务端，通过 HTTP API 下发验证
- 支持 Flag 格式自定义 (`flag{...}`)
- Flag 提交接口限流 (100 req/min/team)

### 3.2 容器层 — Docker 管理

```go
type ContainerManager struct {
    client     *docker.Client
    networkMgr *NetworkManager
    store      ContainerStore
}

type TeamContainer struct {
    ID          string
    TeamID      string
    ChallengeID string
    ContainerID string
    IPAddress   string
    PortMapping map[int]int    // 靶机端口 → 宿主机端口
    Resources   ResourceLimit
    Status      ContainerStatus
}

type ResourceLimit struct {
    CPUCores    float64  // CPU 限制
    MemoryMB    int64    // 内存限制
    DiskGB      int64    // 磁盘限制
    NetworkBPS  int64    // 带宽限制
    PidsLimit   int      // 进程数限制
}

// 核心方法
func (m *ContainerManager) CreateChallenge(ctx context.Context, team *Team, challenge *Challenge) (*TeamContainer, error)
func (m *ContainerManager) DestroyContainer(ctx context.Context, containerID string) error
func (m *ContainerManager) RestartContainer(ctx context.Context, containerID string) error
func (m *ContainerManager) BulkRestart(ctx context.Context, teamIDs []string) error  // 轮间重启
func (m *ContainerManager) MonitorStats(ctx context.Context) ([]ContainerStats, error)
func (m *ContainerManager) EnforceLimits(ctx context.Context, containerID string) error
```

**容器生命周期：**

```
创建镜像 → 分配网络 → 启动容器 → 运行中(监控) → 轮间重启 → 比赛结束销毁
   │           │           │            │              │              │
   │     分配VLAN/      资源限制    CPU/内存监控   保留数据卷      清理网络
   │     分配IP        端口映射    带宽检测      重新部署      释放资源
   │                安全加固
```

### 3.3 网络层

```go
type NetworkManager struct {
    bridgeName string        // 默认 "awd-bridge"
    subnet     string        // 默认 "10.10.0.0/16"
    vlanBase   int           // VLAN 起始 ID
}

// 网络隔离方案
// 1. Docker 自定义桥接网络 (简单场景)
// 2. Open vSwitch + VLAN (高级场景，50+ 队伍)

// 流量镜像
type TrafficMirror struct {
    targetInterface string     // 镜像目标接口
    captureDuration time.Duration
    storagePath     string     // PCAP 存储路径
}

func (n *NetworkManager) CreateTeamNetwork(teamID string) (*docker.NetworkResource, error)
func (n *NetworkManager) IsolateTeams(teamIDs []string) error
func (n *NetworkManager) CaptureTraffic(containerID string, duration time.Duration) (string, error)
func (n *NetworkManager) GetTeamIP(teamID string) (string, error)
```

**网络拓扑：**

```
                        ┌──────────────┐
                        │   宿主机      │
                        │  (awd-br0)   │
                        └──────┬───────┘
                               │
        ┌──────────┬───────────┼───────────┬──────────┐
        ▼          ▼           ▼           ▼          ▼
   ┌─────────┐┌─────────┐┌─────────┐┌─────────┐┌─────────┐
   │ Team 1  ││ Team 2  ││ Team 3  ││ Team 4  ││  ...    │
   │ 10.10.1.││ 10.10.2.││ 10.10.3.││ 10.10.4.││         │
   │ .0/24   ││ .0/24   ││ .0/24   ││ .0/24   ││         │
   └─────────┘└─────────┘└─────────┘└─────────┘└─────────┘
        │                                                     
        ▼                                              流量镜像
   ┌─────────┐                                           │
   │ Port    │◀──────────────────────────────────────────┘
   │ Forward │                                      (afpacket/tc)
   │ :80xx   │
   └─────────┘
        ▼
    (参赛选手)
```

### 3.4 安全层

```go
type SecurityGuard struct {
    waf         *WAFEngine
    ids         *IntrusionDetection
    alertMgr    *AlertManager
    blocklist   *BlocklistManager
}

// WAF 规则引擎
type WAFEngine struct {
    rules       []WAFFilterRule    // SQL注入、XSS、命令注入检测
    rateLimiter *RateLimiter       // 每队请求限流
}

// 流量检测 — 基于 Suricata 规则集
type IntrusionDetection struct {
    ruleSet     string             // 自定义 AWD 规则集
    alertChan   chan *Alert
}

// 异常行为告警
type Alert struct {
    Level       string    // info/warning/critical
    TeamID      string
    Type        string    // "port_scan" | "brute_force" | "exploit" | "suspicious"
    Detail      string
    Timestamp   time.Time
}
```

### 3.5 AI 层 — 智能分析

```
┌────────────────────────────────────────────────────┐
│                  AI Analyzer                        │
│                                                    │
│  ┌──────────────┐  ┌──────────────┐               │
│  │ 攻击模式识别  │  │ 漏洞自动发现  │               │
│  │ (NLP + 规则) │  │ (流量分析)    │               │
│  │              │  │              │               │
│  │ - SQL注入    │  │ - SQLi       │               │
│  │ - XSS        │  │ - RCE        │               │
│  │ - 命令注入   │  │ - SSRF       │               │
│  │ - 文件包含   │  │ - Deser      │               │
│  └──────┬───────┘  └──────┬───────┘               │
│         │                 │                        │
│  ┌──────▼─────────────────▼───────┐               │
│  │        加固建议生成              │               │
│  │  - 代码级修复建议               │               │
│  │  - 配置级加固方案               │               │
│  │  - 优先级排序                   │               │
│  └────────────────────────────────┘               │
└────────────────────────────────────────────────────┘
```

**AI 层实现策略：**
- **初期 (v1.0)**：规则引擎 + 统计分析（无需 GPU，快速落地）
- **中期 (v1.5)**：ONNX Runtime 本地推理（攻击分类模型、异常检测）
- **后期 (v2.0)**：LLM 集成（比赛总结、深度分析报告）

```go
type AIAnalyzer struct {
    ruleEngine    *RuleEngine       // 规则引擎（v1.0）
    mlClassifier  *ONNXClassifier   // ML 分类器（v1.5）
    statsAnalyzer *StatsAnalyzer    // 统计分析
}

type AnalysisReport struct {
    TeamID          string
    AttackPatterns  []AttackPattern
    Vulnerabilities []Vulnerability
    HardeningTips   []HardeningTip
    RiskScore       float64
}
```

### 3.6 监控层

```
┌──────────────────────────────────────────────────┐
│                 Monitor Stack                     │
│                                                  │
│  ┌────────────┐  ┌────────────┐  ┌───────────┐  │
│  │Prometheus  │  │  Grafana   │  │  自定义    │  │
│  │(指标采集)   │  │(可视化)    │  │  大屏      │  │
│  └─────┬──────┘  └────────────┘  └─────┬─────┘  │
│        │                                │        │
│  ┌─────▼────────────────────────────────▼─────┐  │
│  │          WebSocket Push Server              │  │
│  │  - 排名变化实时推送                          │  │
│  │  - 攻击事件实时通知                          │  │
│  │  - 靶机状态变更                              │  │
│  │  - Round 进度倒计时                          │  │
│  └─────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────┘
```

**Prometheus 关键指标：**

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `awd_container_cpu_percent` | Gauge | 各队容器 CPU 使用率 |
| `awd_container_memory_bytes` | Gauge | 各队容器内存使用 |
| `awd_attack_total` | Counter | 攻击提交总数 |
| `awd_attack_success_total` | Counter | 成功攻击数 |
| `awd_round_duration_seconds` | Histogram | Round 处理耗时 |
| `awd_flag_submit_total` | Counter | Flag 提交数 |

---

## 4. 跨平台方案

### 4.1 Windows 兼容方案

| 方案 | 说明 | 推荐度 |
|------|------|--------|
| **WSL2 + Docker Engine** | 最稳定，Linux 原生网络栈 | ⭐⭐⭐⭐⭐ |
| Docker Desktop | 开箱即用，但网络能力受限 | ⭐⭐⭐⭐ |
| 原生 Windows | Go 二进制直接运行 + Docker Desktop | ⭐⭐⭐ |

**推荐 WSL2 方案理由：**
- 网络隔离 (VLAN、iptables) 在 Linux 内核下才能完整支持
- OVS (Open vSwitch) 只能在 Linux 下运行
- PCAP 采集依赖 Linux 内核 af_packet
- Go 二进制在 WSL2 下原生运行

### 4.2 Linux 方案

```bash
# 依赖安装 (Ubuntu/Debian)
apt install -y docker.io openvswitch-switch

# Go 二进制直接运行
./awd-platform server --config config.yaml
```

### 4.3 部署方案

**方案 1：二进制分发（推荐单机）**
```bash
# 编译
GOOS=linux GOARCH=amd64 go build -o awd-platform ./cmd/server
GOOS=windows GOARCH=amd64 go build -o awd-platform.exe ./cmd/server

# 运行
./awd-platform server --config config.yaml
```

**方案 2：Docker Compose（推荐生产）**
```yaml
version: "3.9"
services:
  awd-server:
    build: .
    ports: ["8080:8080", "8443:8443"]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./data:/data
    depends_on:
      - postgres
      - redis
      - nats

  postgres:
    image: postgres:17-alpine
    volumes: ["pgdata:/var/lib/postgresql/data"]
    environment:
      POSTGRES_DB: awd
      POSTGRES_PASSWORD: ${DB_PASSWORD}

  redis:
    image: redis:8-alpine

  nats:
    image: nats:alpine
    command: ["--jetstream", "--store_dir", "/data"]
    volumes: ["natsdata:/data"]

  clickhouse:
    image: clickhouse/clickhouse-server:latest
    volumes: ["chdata:/var/lib/clickhouse"]

volumes:
  pgdata:  natsdata:  chdata:
```

**方案 3：一键部署脚本**

```bash
#!/bin/bash
# deploy.sh — 一键部署
set -e

echo "=== AWD Platform 一键部署 ==="

# 检查 Docker
command -v docker >/dev/null 2>&1 || { echo "请先安装 Docker"; exit 1; }

# 拉取镜像
docker compose pull

# 初始化数据库
docker compose up -d postgres redis nats clickhouse
sleep 5
docker compose exec awd-platform ./awd-platform migrate

# 启动服务
docker compose up -d

echo "=== 部署完成 ==="
echo "管理后台: http://localhost:8080"
echo "默认账号: admin / admin123"
```

---

## 5. 数据模型

### 5.1 ER 关系

```
User 1──N Team N──1 Game 1──N Challenge
     │         │         │         │
     │         │    GameTeam        │
     │         │    (中间表)         │
     │         │         │         │
     ▼         ▼         ▼         ▼
TeamContainer  AttackLog  RoundScore  FlagRecord
```

### 5.2 核心表结构

```sql
-- ==================== 用户与队伍 ====================

CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(64)  NOT NULL UNIQUE,
    password    VARCHAR(256) NOT NULL,  -- bcrypt hash
    email       VARCHAR(128),
    role        VARCHAR(20)  NOT NULL DEFAULT 'player',  -- admin/judge/player
    team_id     BIGINT REFERENCES teams(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE teams (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    token       VARCHAR(64)  NOT NULL UNIQUE,  -- 队伍 API Token
    description TEXT,
    avatar_url  VARCHAR(256),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_team ON users(team_id);

-- ==================== 竞赛 ====================

CREATE TABLE games (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(128) NOT NULL,
    description     TEXT,
    mode            VARCHAR(32)  NOT NULL DEFAULT 'awd_score',  -- awd_score/awd_mix/koh
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',  -- draft/running/paused/finished
    round_duration  INTERVAL     NOT NULL DEFAULT '5 minutes',
    break_duration  INTERVAL     NOT NULL DEFAULT '2 minutes',
    total_rounds    INT          NOT NULL DEFAULT 20,
    current_round   INT          NOT NULL DEFAULT 0,
    current_phase   VARCHAR(20)  NOT NULL DEFAULT 'preparation',
    flag_format     VARCHAR(64)  NOT NULL DEFAULT 'flag{%s}',
    attack_weight   DECIMAL(3,2) NOT NULL DEFAULT 1.0,    -- 攻击得分权重
    defense_weight  DECIMAL(3,2) NOT NULL DEFAULT 0.5,    -- 防守得分权重
    start_time      TIMESTAMPTZ,
    end_time        TIMESTAMPTZ,
    created_by      BIGINT       REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE game_teams (
    game_id     BIGINT REFERENCES games(id),
    team_id     BIGINT REFERENCES teams(id),
    score       DECIMAL(10,2) NOT NULL DEFAULT 0,
    rank        INT,
    PRIMARY KEY (game_id, team_id)
);

CREATE INDEX idx_games_status ON games(status);

-- ==================== 靶机与容器 ====================

CREATE TABLE challenges (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT       REFERENCES games(id),
    name        VARCHAR(128) NOT NULL,
    description TEXT,
    image_name  VARCHAR(256) NOT NULL,     -- Docker 镜像名
    image_tag   VARCHAR(64)  DEFAULT 'latest',
    difficulty  VARCHAR(20)  DEFAULT 'medium',  -- easy/medium/hard
    base_score  INT          NOT NULL DEFAULT 100,
    exposed_ports JSONB,    -- [{"container": 80, "protocol": "tcp"}]
    cpu_limit   DECIMAL(3,2) DEFAULT 0.5,   -- CPU 核心数
    mem_limit   INT          DEFAULT 256,    // MB
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE team_containers (
    id              BIGSERIAL PRIMARY KEY,
    game_id         BIGINT REFERENCES games(id),
    team_id         BIGINT REFERENCES teams(id),
    challenge_id    BIGINT REFERENCES challenges(id),
    container_id    VARCHAR(128),            -- Docker Container ID
    ip_address      VARCHAR(45),             -- IPv4/IPv6
    port_mapping    JSONB,                   -- {"80": 18001, "22": 22001}
    status          VARCHAR(20)  NOT NULL DEFAULT 'creating',  -- creating/running/stopped/error
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, team_id, challenge_id)
);

-- ==================== Flag 与提交 ====================

CREATE TABLE flag_records (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT REFERENCES games(id),
    round       INT          NOT NULL,
    team_id     BIGINT REFERENCES teams(id),     -- 被攻击的队伍
    flag_hash   VARCHAR(256) NOT NULL,           -- Flag 的哈希 (防泄露)
    flag_value  VARCHAR(256) NOT NULL,           -- Flag 明文
    service     VARCHAR(128) NOT NULL,           -- 服务名称
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, round, team_id, service)
);

CREATE TABLE flag_submissions (
    id              BIGSERIAL PRIMARY KEY,
    game_id         BIGINT REFERENCES games(id),
    round           INT          NOT NULL,
    attacker_team   BIGINT REFERENCES teams(id),
    target_team     BIGINT REFERENCES teams(id),
    flag_value      VARCHAR(256) NOT NULL,
    is_correct      BOOLEAN      NOT NULL,
    points_earned   DECIMAL(10,2) NOT NULL DEFAULT 0,
    submitted_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, round, attacker_team, target_team, flag_value)
);

CREATE INDEX idx_submissions_game_round ON flag_submissions(game_id, round);

-- ==================== 积分 ====================

CREATE TABLE round_scores (
    id              BIGSERIAL PRIMARY KEY,
    game_id         BIGINT REFERENCES games(id),
    round           INT          NOT NULL,
    team_id         BIGINT REFERENCES teams(id),
    attack_score    DECIMAL(10,2) NOT NULL DEFAULT 0,  -- 攻击得分
    defense_score   DECIMAL(10,2) NOT NULL DEFAULT 0,  -- 防守得分 (扣分)
    total_score     DECIMAL(10,2) NOT NULL DEFAULT 0,
    rank            INT,
    calculated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (game_id, round, team_id)
);

-- ==================== 攻击日志 (ClickHouse) ====================

CREATE TABLE attack_logs (
    timestamp       DateTime64(3),
    game_id         UInt64,
    round           UInt32,
    attacker_team   String,
    target_team     String,
    target_ip       String,
    target_port     UInt16,
    protocol        String,         -- tcp/udp/http
    method          Nullable(String),  -- HTTP method
    path            Nullable(String),  -- URL path
    payload_hash    String,
    attack_type     String,         -- sql_injection/xss/rce/...
    severity        String,         -- low/medium/high/critical
    source_ip       String,
    user_agent      Nullable(String),
    raw_log         String
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (game_id, round, attacker_team, timestamp);

-- ==================== 事件日志 ====================

CREATE TABLE event_logs (
    id          BIGSERIAL PRIMARY KEY,
    game_id     BIGINT,
    event_type  VARCHAR(64)  NOT NULL,  -- round_start/attack/defense/alert/...
    level       VARCHAR(20)  NOT NULL DEFAULT 'info',
    team_id     BIGINT,
    detail      JSONB,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_game_type ON event_logs(game_id, event_type);
```

---

## 6. API 设计

### 6.1 RESTful API 端点

#### 认证
| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/logout` | 登出 |
| GET | `/api/v1/auth/me` | 当前用户信息 |

#### 用户管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/users` | 用户列表 (Admin) |
| POST | `/api/v1/users` | 创建用户 |
| PUT | `/api/v1/users/:id` | 更新用户 |
| DELETE | `/api/v1/users/:id` | 删除用户 |

#### 队伍管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/teams` | 队伍列表 |
| POST | `/api/v1/teams` | 创建队伍 |
| GET | `/api/v1/teams/:id` | 队伍详情 |
| GET | `/api/v1/teams/:id/members` | 队伍成员 |

#### 竞赛管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/games` | 竞赛列表 |
| POST | `/api/v1/games` | 创建竞赛 |
| GET | `/api/v1/games/:id` | 竞赛详情 |
| PUT | `/api/v1/games/:id` | 更新竞赛 |
| POST | `/api/v1/games/:id/start` | 开始竞赛 |
| POST | `/api/v1/games/:id/pause` | 暂停竞赛 |
| POST | `/api/v1/games/:id/stop` | 结束竞赛 |
| POST | `/api/v1/games/:id/reset` | 重置竞赛 |

#### 靶机/容器
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/challenges` | 靶机列表 |
| POST | `/api/v1/games/:id/challenges` | 添加靶机 |
| GET | `/api/v1/games/:id/containers` | 容器状态 |
| POST | `/api/v1/games/:id/containers/restart` | 重启所有容器 |
| POST | `/api/v1/games/:id/containers/:cid/restart` | 重启单个容器 |
| GET | `/api/v1/games/:id/containers/stats` | 容器资源统计 |

#### Flag 提交
| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/games/:id/flags/submit` | 提交 Flag |
| GET | `/api/v1/games/:id/flags/history` | 提交历史 |

#### 排名
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/rankings` | 实时排名 |
| GET | `/api/v1/games/:id/rankings/rounds/:round` | 指定轮次排名 |

#### 安全事件
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/alerts` | 安全告警列表 |
| GET | `/api/v1/games/:id/attacks` | 攻击日志 |

#### AI 分析
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/ai/team/:teamId/report` | 队伍分析报告 |
| GET | `/api/v1/games/:id/ai/summary` | 比赛总体分析 |

### 6.2 WebSocket 事件

**连接**: `ws://host:8080/ws?token=JWT`

```json
// 客户端 → 服务端
{ "type": "subscribe", "channel": "game:123" }
{ "type": "unsubscribe", "channel": "game:123" }

// 服务端 → 客户端
{ "type": "round:start", "data": { "round": 5, "ends_at": "...", "phase": "running" } }
{ "type": "round:end", "data": { "round": 5, "rankings": [...] } }
{ "type": "ranking:update", "data": { "rankings": [...] } }
{ "type": "flag:captured", "data": { "attacker": "Team A", "target": "Team B", "points": 100 } }
{ "type": "alert:new", "data": { "level": "warning", "team": "Team C", "message": "..." } }
{ "type": "container:status", "data": { "team_id": "1", "status": "running", "cpu": 45.2 } }
{ "type": "game:status", "data": { "status": "paused", "reason": "..." } }
```

### 6.3 API 通用响应格式

```typescript
// 成功
interface APIResponse<T> {
    code: number;        // 0
    message: string;     // "ok"
    data: T;
}

// 分页
interface PagedResponse<T> extends APIResponse<T[]> {
    pagination: {
        page: number;
        page_size: number;
        total: number;
        total_pages: number;
    };
}

// 错误
interface APIError {
    code: number;        // 非 0
    message: string;     // 错误描述
    details?: any;       // 详细信息
}
```

---

## 7. 部署架构

### 7.1 单机部署（开发/小型比赛 ≤ 20 队）

```
┌──────────────────────────────────────────┐
│              单台服务器 (8核16G+)          │
│                                          │
│  ┌──────────────────────────────────┐    │
│  │  Docker Compose                  │    │
│  │  ┌─────────┐ ┌────────────────┐ │    │
│  │  │ AWD     │ │ PostgreSQL     │ │    │
│  │  │ Server  │ │ Redis          │ │    │
│  │  │ (Go)    │ │ NATS           │ │    │
│  │  └────┬────┘ │ ClickHouse     │ │    │
│  │       │      └────────────────┘ │    │
│  │  ┌────▼────────────────────┐    │    │
│  │  │  Docker-in-Docker       │    │    │
│  │  │  (50+ 靶机容器)          │    │    │
│  │  └─────────────────────────┘    │    │
│  └──────────────────────────────────┘    │
└──────────────────────────────────────────┘
```

**资源需求**: 8核 CPU / 16GB RAM / 100GB SSD

### 7.2 分布式部署（大型比赛 50+ 队）

```
┌─────────────────────────────────────────────────────────────┐
│                         负载均衡 (Nginx)                      │
│                    :80 / :443 / :8443                        │
└──────────┬──────────────────────────┬────────────────────────┘
           │                          │
           ▼                          ▼
┌──────────────────┐      ┌──────────────────┐
│   Web 节点 1     │      │   Web 节点 2     │
│   (API + 前端)   │      │   (API + 前端)   │
└───────┬──────────┘      └───────┬──────────┘
        │                        │
        ▼                        ▼
┌────────────────────────────────────────────┐
│              NATS Cluster (3 节点)           │
│         事件总线 + 状态同步                   │
└──────────────────┬─────────────────────────┘
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
┌────────────┐ ┌────────┐ ┌────────────┐
│PostgreSQL  │ │ Redis  │ │ClickHouse  │
│(主从)      │ │(Sentinel)│ │(集群)     │
└────────────┘ └────────┘ └────────────┘

┌────────────────────────────────────────────┐
│           Docker 宿主机集群                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Host 1   │  │ Host 2   │  │ Host 3   │ │
│  │ ~17队    │  │ ~17队    │  │ ~17队    │ │
│  │ 容器     │  │ 容器     │  │ 容器     │ │
│  └──────────┘  └──────────┘  └──────────┘ │
└────────────────────────────────────────────┘

┌────────────────────────────────────────────┐
│              监控集群                        │
│  Prometheus + Grafana + AlertManager       │
└────────────────────────────────────────────┘
```

**资源需求（每台）**: 16核 CPU / 32GB RAM / 500GB SSD × 3

### 7.3 关键配置

```yaml
# config.yaml
server:
  host: 0.0.0.0
  port: 8080
  tls_port: 8443
  jwt_secret: "${JWT_SECRET}"

database:
  postgres: "postgres://awd:password@localhost:5432/awd?sslmode=disable"
  redis: "redis://localhost:6379/0"
  clickhouse: "clickhouse://localhost:9000/awd"

nats:
  url: "nats://localhost:4222"

docker:
  host: "unix:///var/run/docker.sock"   # Linux
  # host: "npipe:////./pipe/docker_engine"  # Windows

security:
  jwt_expire_hours: 24
  rate_limit: 100          # req/min per team
  flag_submit_limit: 100   # flag submits/min per team
  blocklist_enabled: true

game:
  default_round_duration: 5m
  default_break_duration: 2m
  max_containers_per_team: 5

ai:
  enabled: true
  rule_engine_path: "./rules/"
  onnx_model_path: "./models/"
```

---

## 8. 项目目录结构

```
awd-platform/
├── cmd/                            # 入口
│   ├── server/                     # 主服务
│   │   └── main.go
│   ├── cli/                        # 命令行工具
│   │   └── main.go
│   └── migrator/                   # 数据库迁移
│       └── main.go
│
├── internal/                       # 内部包（不对外暴露）
│   ├── config/                     # 配置管理
│   │   └── config.go
│   ├── middleware/                  # HTTP/gRPC 中间件
│   │   ├── auth.go                 # JWT 认证
│   │   ├── ratelimit.go            # 限流
│   │   ├── logger.go               # 日志
│   │   └── cors.go                 # CORS
│   │
│   ├── server/                     # HTTP 服务器
│   │   ├── server.go
│   │   ├── router.go
│   │   └── ws.go                   # WebSocket
│   │
│   ├── engine/                     # 🔥 竞赛引擎
│   │   ├── engine.go               # 引擎主逻辑
│   │   ├── round.go                # Round 调度
│   │   ├── scoring.go              # 评分计算
│   │   ├── flag.go                 # Flag 管理
│   │   └── mode/                   # 竞赛模式
│   │       ├── mode.go             # 模式接口
│   │       ├── awd_score.go        # AWD 经典
│   │       ├── awd_mix.go          # 攻防混合
│   │       └── koh.go              # 山顶争夺
│   │
│   ├── container/                  # 🔥 Docker 容器管理
│   │   ├── manager.go              # 容器生命周期
│   │   ├── monitor.go              # 资源监控
│   │   ├── limits.go               # 资源限制
│   │   └── image.go                # 镜像管理
│   │
│   ├── network/                    # 🔥 网络管理
│   │   ├── manager.go              # 网络创建/隔离
│   │   ├── mirror.go               # 流量镜像
│   │   ├── capture.go              # PCAP 采集
│   │   └── ovs.go                  # OVS VLAN (Linux)
│   │
│   ├── security/                   # 🔥 安全层
│   │   ├── waf.go                  # WAF 规则引擎
│   │   ├── ids.go                  # 入侵检测
│   │   ├── alert.go                # 告警管理
│   │   └── rules/                  # 规则文件
│   │       ├── sql_injection.yaml
│   │       ├── xss.yaml
│   │       └── command_injection.yaml
│   │
│   ├── ai/                         # 🔥 AI 分析层
│   │   ├── analyzer.go             # 分析入口
│   │   ├── rule_engine.go          # 规则引擎
│   │   ├── stats.go                # 统计分析
│   │   ├── classifier.go           # ML 分类器 (v1.5)
│   │   └── report.go               # 报告生成
│   │
│   ├── monitor/                    # 🔥 监控层
│   │   ├── metrics.go              # Prometheus 指标
│   │   ├── health.go               # 健康检查
│   │   └── push.go                 # WebSocket 推送
│   │
│   ├── model/                      # 数据模型
│   │   ├── user.go
│   │   ├── team.go
│   │   ├── game.go
│   │   ├── challenge.go
│   │   ├── container.go
│   │   ├── flag.go
│   │   ├── score.go
│   │   └── attack.go
│   │
│   ├── repo/                       # 数据访问层
│   │   ├── postgres/               # PostgreSQL 实现
│   │   │   ├── user_repo.go
│   │   │   ├── game_repo.go
│   │   │   ├── flag_repo.go
│   │   │   └── score_repo.go
│   │   ├── redis/                  # Redis 实现
│   │   │   ├── ranking_repo.go
│   │   │   └── cache_repo.go
│   │   └── clickhouse/             # ClickHouse 实现
│   │       └── attack_log_repo.go
│   │
│   ├── service/                    # 业务逻辑层
│   │   ├── auth_service.go
│   │   ├── game_service.go
│   │   ├── team_service.go
│   │   ├── flag_service.go
│   │   ├── container_service.go
│   │   ├── ranking_service.go
│   │   └── ai_service.go
│   │
│   ├── handler/                    # HTTP handler
│   │   ├── auth_handler.go
│   │   ├── game_handler.go
│   │   ├── team_handler.go
│   │   ├── flag_handler.go
│   │   ├── container_handler.go
│   │   ├── ranking_handler.go
│   │   └── ai_handler.go
│   │
│   └── eventbus/                   # 事件总线
│       ├── bus.go                  # NATS 封装
│       ├── events.go               # 事件定义
│       └── handler.go              # 事件处理器
│
├── pkg/                            # 可导出的公共包
│   ├── logger/                     # 日志 (slog 封装)
│   ├── validator/                  # 参数校验
│   ├── crypto/                     # 加密工具
│   └── httputil/                   # HTTP 工具
│
├── api/                            # API 定义
│   └── openapi.yaml                # OpenAPI 3.1 规范
│
├── migrations/                     # 数据库迁移
│   ├── 001_init_schema.sql
│   ├── 002_seed_data.sql
│   └── 003_clickhouse.sql
│
├── web/                            # 前端
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── api/                    # API 调用
│   │   ├── components/             # 通用组件
│   │   ├── pages/                  # 页面
│   │   │   ├── Dashboard/          # 大屏
│   │   │   ├── GameManage/         # 竞赛管理
│   │   │   ├── Ranking/            # 排名
│   │   │   ├── TeamManage/         # 队伍管理
│   │   │   └── Settings/           # 设置
│   │   ├── hooks/                  # 自定义 Hooks
│   │   ├── stores/                 # 状态管理 (Zustand)
│   │   └── utils/
│   └── public/
│
├── configs/                        # 配置文件模板
│   ├── config.yaml
│   ├── config.dev.yaml
│   └── config.prod.yaml
│
├── scripts/                        # 工具脚本
│   ├── deploy.sh                   # 一键部署
│   ├── build.sh                    # 构建
│   └── setup-dev.sh               # 开发环境搭建
│
├── deployments/                    # 部署配置
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── Dockerfile.frontend
│   ├── docker-compose.yml
│   └── docker-compose.prod.yml
│
├── docs/                           # 文档
│   ├── 01-requirements.md
│   ├── 02-architecture.md          ← 本文档
│   ├── 03-api-spec.md
│   └── 04-operation-guide.md
│
├── tests/                          # 测试
│   ├── unit/                       # 单元测试
│   ├── integration/                # 集成测试
│   └── e2e/                        # E2E 测试
│
├── .golangci.yml                   # Go Lint
├── .gitignore
├── Makefile
├── go.mod
└── README.md
```

---

## 附录

### A. Go 依赖清单

```
github.com/gofiber/fiber/v3          # Web 框架
github.com/golang-jwt/jwt/v5         # JWT
github.com/docker/docker             # Docker SDK
github.com/jackc/pgx/v5             # PostgreSQL driver
github.com/redis/go-redis/v9         # Redis client
github.com/ClickHouse/clickhouse-go  # ClickHouse driver
github.com/nats-io/nats.go           # NATS client
github.com/casbin/casbin/v2          # 权限控制
github.com/go-playground/validator   # 参数校验
github.com/prometheus/client_golang  # Prometheus
github.com/gorilla/websocket         # WebSocket
github.com/minio/minio-go/v7         # MinIO
github.com/yuin/gopher-lua           # 规则引擎 (Lua)
github.com/onnxruntime-go/onnxruntime  # ONNX (v1.5)
```

### B. 前端依赖清单

```json
{
  "dependencies": {
    "react": "^19",
    "react-dom": "^19",
    "react-router": "^7",
    "zustand": "^5",
    "axios": "^1",
    "recharts": "^2",
    "dayjs": "^1",
    "@tanstack/react-query": "^5",
    "antd": "^5"
  },
  "devDependencies": {
    "typescript": "^5",
    "vite": "^6",
    "@vitejs/plugin-react": "^4",
    "tailwindcss": "^4",
    "eslint": "^9"
  }
}
```

### C. 性能预估

| 指标 | 目标值 |
|------|--------|
| 并发队伍 | 50+ |
| 每轮 Flag 提交峰值 | 1000 req/s |
| WebSocket 连接数 | 200+ |
| 容器启动时间 | < 3s/个 |
| 排行榜更新延迟 | < 500ms |
| 攻击日志写入吞吐 | 10k events/s |

---

> **下一步**: 基于本架构，可进入详细 API 设计 (`03-api-spec.md`) 和数据库 Migration 编写。
