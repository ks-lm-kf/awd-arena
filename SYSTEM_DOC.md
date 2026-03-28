# AWD Arena 系统文档

> **最后更新**: 2026-03-28
> **版本**: v0.1.0 (Initial)
> **本文档面向**: AI 开发助手 / 新开发者
> **项目路径**: `/opt/awd-arena`
> **技术栈**: Go 1.25 + Fiber v3 + GORM + SQLite + Docker + React

---

## 目录

1. [项目概述](#1-项目概述)
2. [目录结构](#2-目录结构)
3. [技术架构](#3-技术架构)
4. [数据模型](#4-数据模型)
5. [API 接口](#5-api-接口)
6. [核心引擎](#6-核心引擎)
7. [计分系统](#7-计分系统)
8. [比赛状态机](#8-比赛状态机)
9. [轮次管理](#9-轮次管理)
10. [Flag 系统](#10-flag-系统)
11. [权限系统 (RBAC)](#11-权限系统-rbac)
12. [中间件](#12-中间件)
13. [WebSocket](#13-websocket)
14. [容器管理](#14-容器管理)
15. [安全 & WAF](#15-安全--waf)
16. [通知系统](#16-通知系统)
17. [导出功能](#17-导出功能)
18. [AI 分析](#18-ai-分析)
19. [配置](#19-配置)
20. [数据库](#20-数据库)
21. [前端](#21-前端)
22. [开发指南 & 约定](#22-开发指南--约定)
23. [更新日志](#23-更新日志)

---

## 1. 项目概述

AWD Arena 是一个 **AWD (Attack With Defense) 竞技平台**，用于举办网络安全攻防对抗比赛。

### 核心功能
- **比赛管理**: 创建、启动、暂停、恢复、停止 AWD 比赛
- **自动轮次**: 可配置轮次时长、休息时长、总轮数
- **Flag 系统**: 每轮自动生成 Flag，写入各队容器
- **零和计分**: 攻击得分 = 被攻方失分，保持总分恒定
- **防御奖励**: 未被攻破的队伍获得额外防御分
- **首杀奖励**: 每道题第一个攻破的队伍获得额外加分
- **容器管理**: 基于 Docker 为每支队伍自动创建题目容器
- **实时排名**: WebSocket 推送排名变化
- **RBAC 权限**: 管理员 / 裁判 / 选手 三级权限
- **WAF 安全**: 内置 Web 应用防火墙
- **通知系统**: 飞书 / 邮件 / WebSocket 通知

### 默认管理员
- 用户名: `admin` / 密码: `admin123` (首次启动自动创建)

---

## 2. 目录结构

```
/opt/awd-arena/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口，初始化数据库、引擎回调、HTTP服务
├── internal/
│   ├── config/
│   │   └── config.go            # 配置加载（YAML）
│   ├── database/
│   │   └── database.go          # SQLite 数据库初始化
│   ├── model/                   # 数据模型（见第4节）
│   │   ├── game.go              # Game, GameTeam
│   │   ├── user.go              # User
│   │   ├── team.go              # Team, GameTeam
│   │   ├── challenge.go         # Challenge (题目)
│   │   ├── flag.go              # FlagRecord, FlagSubmission
│   │   ├── score.go             # RoundScore
│   │   ├── container.go         # TeamContainer
│   │   ├── leaderboard.go       # LeaderboardEntry, ScoreUpdate
│   │   ├── ranking.go           # Ranking, RoundRanking, CompetitionStats
│   │   ├── attack.go            # AttackLog, EventLog
│   │   ├── docker_image.go      # DockerImage
│   │   ├── challenge_template.go # ChallengeTemplate + ImageConfig/VulnConfig/FlagConfig
│   │   ├── service_health.go    # ServiceHealth, HealthCheckConfig
│   │   ├── target_service.go    # TargetService
│   │   ├── score_adjustment.go  # ScoreAdjustment
│   │   ├── admin_log.go         # AdminLog
│   │   └── rbac_permissions.go  # Role, Permission, RolePermissions
│   ├── engine/                  # 核心引擎（见第6-10节）
│   │   ├── engine.go            # CompetitionEngine 主引擎
│   │   ├── scoring.go           # ScoreCalculator 简化计分
│   │   ├── manager.go           # EngineManager 全局引擎管理器
│   │   ├── game_state_machine.go # GameStateMachine 状态机
│   │   ├── round.go             # RoundScheduler 轮次调度器
│   │   ├── round_manager.go     # RoundManager 高级轮次管理
│   │   ├── flag.go              # FlagManager Flag管理
│   │   ├── flag_generator.go    # FlagGenerator Flag生成器
│   │   ├── flag_writer.go       # FlagWriter Docker容器Flag写入
│   │   └── scoring/             # 高级计分模块
│   │       ├── zero_sum.go      # ZeroSumScorer 零和计分
│   │       ├── defense.go       # DefenseCalculator 防御分
│   │       └── first_blood.go   # FirstBloodDetector 首杀检测
│   ├── service/                 # 业务服务层
│   │   ├── game_service.go      # GameService
│   │   ├── score_service.go     # ScoreService
│   │   ├── flag_service.go      # FlagService
│   │   ├── team_service.go      # TeamService
│   │   ├── container_service.go # ContainerService
│   │   ├── challenge_service.go # ChallengeService
│   │   ├── auth_service.go      # AuthService (JWT认证)
│   │   ├── ranking_service.go   # RankingService
│   │   ├── ai_service.go        # AIService
│   │   └── docker_image_service.go # DockerImageService
│   ├── handler/                 # HTTP 请求处理器
│   │   ├── admin_handler.go     # 管理操作（带审计日志）
│   │   ├── auth_handler.go      # 登录/注册/刷新Token
│   │   ├── dashboard_handler.go # 仪表盘
│   │   ├── flag_handler.go      # Flag 提交/历史
│   │   ├── leaderboard_handler.go # 排行榜
│   │   ├── round_handler.go     # 轮次查询/控制
│   │   ├── score_handler.go     # 分数查询
│   │   ├── security_handler.go  # 安全相关
│   │   ├── settings_handler.go  # 系统设置
│   │   ├── template_handler.go  # 题目模板
│   │   ├── waf_handler.go       # WAF 规则
│   │   ├── audit_handler.go     # 审计日志
│   │   ├── export_handler.go    # 数据导出
│   │   └── docker_image_handler.go # Docker镜像管理
│   ├── repo/                    # 数据仓库层
│   │   ├── flag_repo.go         # FlagRepo 接口
│   │   ├── game_repo.go         # GameRepo 接口
│   │   ├── score_repo.go        # ScoreRepo 接口
│   │   ├── postgres/            # PostgreSQL 实现 (预留)
│   │   ├── redis/               # Redis 缓存/排名
│   │   └── clickhouse/          # ClickHouse 攻击日志
│   ├── server/                  # HTTP 服务
│   │   ├── server.go            # Fiber 服务器初始化
│   │   ├── router.go            # 路由注册（见第5节）
│   │   └── ws.go                # WebSocket Hub
│   ├── middleware/              # 中间件
│   │   ├── auth.go              # JWT 认证
│   │   ├── rbac_auth.go         # RBAC 角色/权限检查
│   │   ├── rbac_authorizer.go   # RBAC 授权器
│   │   ├── permission.go        # 权限定义
│   │   ├── audit.go             # 审计日志
│   │   ├── cors.go              # CORS
│   │   ├── logger.go            # 请求日志
│   │   ├── ratelimit.go         # 速率限制
│   │   ├── security_headers.go  # 安全头
│   │   ├── waf_middleware.go     # WAF 中间件
│   │   └── resource_auth.go     # 资源级权限
│   ├── container/               # Docker 容器管理
│   │   ├── docker_client.go     # Docker 客户端
│   │   ├── manager.go           # 容器生命周期管理
│   │   ├── limits.go            # 资源限制
│   │   ├── gorm_store.go        # 容器持久化
│   │   └── image_extensions.go  # 镜像扩展
│   ├── security/                # 安全模块
│   │   ├── waf.go               # Web应用防火墙
│   │   ├── alert.go             # 安全告警
│   │   ├── attack_log.go        # 攻击日志
│   │   └── utils.go             # 安全工具
│   ├── eventbus/                # 事件总线
│   │   ├── bus.go               # 事件发布/订阅
│   │   ├── events.go            # 事件定义
│   │   ├── broadcaster.go       # 事件广播
│   │   └── handler.go           # 事件处理器
│   ├── notify/                  # 通知模块
│   │   ├── feishu.go            # 飞书通知
│   │   ├── email.go             # 邮件通知
│   │   └── websocket.go         # WebSocket 推送
│   ├── network/                 # 网络管理
│   │   └── manager.go           # 网络隔离/管理
│   ├── ai/                      # AI 分析模块
│   │   ├── analyzer.go          # 攻击分析
│   │   ├── classifier.go        # 攻击分类
│   │   ├── report.go            # 报告生成
│   │   └── rule_engine.go       # 规则引擎
│   ├── export/                  # 数据导出
│   │   ├── csv.go               # CSV 导出
│   │   └── pdf.go               # PDF 导出
│   └── monitor/                 # 监控
│       ├── health_checker.go    # 健康检查
│       ├── health.go            # 健康状态
│       ├── metrics.go           # Prometheus 指标
│       └── push.go              # 推送通知
├── pkg/                         # 公共包
│   ├── crypto/                  # 加密工具（密码哈希、Flag生成）
│   └── logger/                  # 日志工具
├── web/                         # React 前端 (SPA)
└── configs/
    └── config.yaml              # 配置文件
```

---

## 3. 技术架构

### 架构图 (文本)

```
┌─────────────────────────────────────────────────────────┐
│                     React SPA (前端)                      │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP / WebSocket
┌────────────────────────▼────────────────────────────────┐
│                  Fiber v3 HTTP Server                     │
│  ┌──────────────────────────────────────────────────┐   │
│  │               Middleware Chain                     │   │
│  │  SecurityHeaders → CORS → Logger → WAF → RateLimit│   │
│  │  → JWTAuth → RBAC → Audit → ResourceAuth          │   │
│  └──────────────────────────────────────────────────┘   │
│  ┌────────────┐  ┌────────────┐  ┌─────────────────┐   │
│  │  Handler   │  │  Handler   │  │   Handler ...   │   │
│  └─────┬──────┘  └─────┬──────┘  └────────┬────────┘   │
│        └────────────────┼─────────────────┘            │
│                   ┌─────▼──────┐                        │
│                   │  Service   │  (业务逻辑层)           │
│                   └─────┬──────┘                        │
│          ┌──────────────┼──────────────┐                │
│    ┌─────▼──────┐ ┌─────▼──────┐ ┌─────▼──────┐        │
│    │   Repo     │ │  Engine    │ │ Container  │        │
│    │ (数据访问) │ │ (核心引擎) │ │ (Docker)   │        │
│    └─────┬──────┘ └─────┬──────┘ └─────┬──────┘        │
└──────────┼──────────────┼──────────────┼───────────────┘
           │              │              │
    ┌──────▼──────┐ ┌─────▼──────┐ ┌─────▼──────┐
    │   SQLite    │ │ EventBus   │ │   Docker    │
    │  (主数据库)  │ │ (事件总线)  │ │  (容器引擎) │
    └─────────────┘ └─────┬──────┘ └────────────┘
                          │
                   ┌──────▼──────┐
                   │  WebSocket  │
                   │  (实时推送)  │
                   └─────────────┘
```

### 分层架构
- **Handler** → 接收 HTTP 请求，参数校验，调用 Service
- **Service** → 业务逻辑，编排调用
- **Repo** → 数据访问抽象（接口 + 实现）
- **Engine** → 比赛核心引擎（状态机、轮次、计分）
- **Model** → GORM 数据模型

### 依赖关系
- `cmd/server/main.go` 初始化一切：DB → AutoMigrate → EngineCallbacks → Server
- `EngineManager` (全局单例) 管理所有活跃的 `CompetitionEngine`
- `service.EngineCallbacks` 桥接 Service 层和 Engine 层，避免循环依赖

---

## 4. 数据模型

### 4.1 User (用户)
```go
type User struct {
    ID                 int64      // 主键
    Username           string     // 唯一索引
    Password           string     // bcrypt 哈希 (JSON 不输出)
    Email              string
    Role               string     // admin / organizer / player
    TeamID             *int64     // 所属队伍
    PasswordChangedAt  *time.Time
    MustChangePassword bool       // 首次登录强制改密码
    CreatedAt, UpdatedAt time.Time
}
```

### 4.2 Team (队伍)
```go
type Team struct {
    ID          int64     // 主键
    Name        string    // 唯一索引
    Token       string    // 队伍 Token (唯一, API 用)
    Description string
    AvatarURL   string
    Score       float64   // 当前总分
    CreatedAt   time.Time
}

type GameTeam struct {  // 比赛-队伍关联
    ID, GameID, TeamID int64
    Team               Team  // 外键
}
```

### 4.3 Game (比赛)
```go
type Game struct {
    ID             int64      // 主键
    Title          string
    Description    string
    Mode           string     // 默认 "awd_score"
    Status         string     // draft / active / finished
    CurrentPhase   string     // preparation / running / break / finished
    RoundDuration  int        // 轮次时长(秒), 默认 300
    BreakDuration  int        // 休息时长(秒), 默认 120
    TotalRounds    int        // 总轮数, 默认 20
    CurrentRound   int        // 当前轮次
    FlagFormat     string     // Flag格式, 默认 "flag{%s}"
    AttackWeight   float64    // 攻击权重, 默认 1.0
    DefenseWeight  float64    // 防守权重, 默认 0.5
    StartTime      *time.Time
    EndTime        *time.Time
    CreatedBy      int64
    CreatedAt, UpdatedAt time.Time
}
// 方法: IsDraft(), IsActive(), IsFinished(), CanStart(), CanPause(), CanResume(), CanFinish()
```

**状态流转**:
```
draft (preparation) ──start──▶ active (running) ──pause──▶ active (break)
                                     ▲                       │
                                     └─────resume─────────────┘
                                     │
                                  finish
                                     ▼
                              finished (finished)
```

### 4.4 Challenge (题目)
```go
type Challenge struct {
    ID           int64
    GameID       int64      // 所属比赛
    Name         string
    Description  string
    ImageName    string     // Docker 镜像名
    ImageTag     string     // 默认 "latest"
    Difficulty   string     // easy / medium / hard
    BaseScore    int        // 默认 100
    ExposedPorts string     // JSON: 端口映射
    CPULimit     float64    // 默认 0.5 核
    MemLimit     int        // 默认 256 MB
    CreatedAt    time.Time
}
```

### 4.5 Flag 相关
```go
type FlagRecord struct {       // 生成的 Flag
    ID        int64
    GameID    int64            // 复合索引: idx_game_round_team
    Round     int
    TeamID    int64
    FlagHash  string           // SHA256
    FlagValue string           // 明文 (JSON 不输出)
    Service   string
    CreatedAt time.Time
}

type FlagSubmission struct {   // Flag 提交记录
    ID           int64
    GameID, Round int
    AttackerTeam  int64        // 攻击方
    TargetTeam    int64        // 被攻方
    FlagValue     string
    IsCorrect     bool
    PointsEarned  float64
    SubmittedAt   time.Time
}
```

### 4.6 Score (分数)
```go
type RoundScore struct {       // 每轮每队分数
    ID           int64
    GameID       int64
    Round        int
    TeamID       int64
    AttackScore  float64       // 攻击分
    DefenseScore float64       // 防御分 (负数=被扣分, 正数=奖励)
    TotalScore   float64       // 总分
    Rank         int           // 排名
    CalculatedAt time.Time
}

type ScoreAdjustment struct {  // 人工调分
    ID          int64
    GameID      int64
    TeamID      int64
    AdjustValue int            // 正=加分, 负=减分
    Reason      string
    OperatorID  int64
    Round       int
    CreatedAt   time.Time
}
```

### 4.7 TeamContainer (容器)
```go
type TeamContainer struct {
    ID          int64
    GameID      int64
    TeamID      int64
    ChallengeID int64
    ContainerID string         // Docker 容器 ID
    IPAddress   string
    PortMapping string         // JSON
    SSHUser     string         // 默认 "awd"
    SSHPassword string
    Status      string         // creating / running / stopped / error
    CreatedAt   time.Time
}
```

### 4.8 Leaderboard / Ranking
```go
type LeaderboardEntry struct {
    Rank                    int
    TeamID                  int64
    TeamName                string
    TotalScore, AttackScore float64
    DefenseScore            float64
    FirstBloods             int
}
type ScoreUpdate struct { GameID int64; Entries []LeaderboardEntry }

type Ranking struct {            // 缓存排名
    Rank, TeamID uint; TeamName string; Score float64
    Attacks, Defenses, FirstBlood int
}
```

### 4.9 AttackLog / EventLog
```go
type AttackLog struct {          // 攻击日志 (ClickHouse)
    Timestamp    time.Time
    GameID       uint64; Round uint32
    AttackerTeam, TargetTeam string
    TargetIP     string; TargetPort uint16
    Protocol     string
    Method, Path *string
    PayloadHash  string
    AttackType   string; Severity string
    SourceIP     string
    UserAgent    *string
    RawLog       string
}
type EventLog struct {           // 系统事件
    ID int64; GameID *int64; EventType, Level string
    TeamID *int64; Detail string  // JSON
    CreatedAt time.Time
}
```

### 4.10 DockerImage
```go
type DockerImage struct {
    ID           uint
    Name, Tag    string
    ImageID      string
    Description  string
    Category     string         // general / web / pwn / ...
    Difficulty   string
    Ports        string
    MemoryLimit  int            // MB
    CPULimit     float64
    Flag         string
    InitialScore int
    Status       string         // active / inactive
    CreatedAt, UpdatedAt time.Time
}
```

### 4.11 ChallengeTemplate (题目模板)
```go
type ChallengeTemplate struct {
    ID          int64; Name string  // 唯一索引
    Category    string              // web / pwn / crypto / misc / reverse
    Description string
    ImageConfig   ImageConfig       // JSON: 镜像名/标签/仓库/环境变量/卷挂载/网络模式
    ServicePorts  ServicePorts      // JSON: 端口/协议/服务名/描述
    VulnConfig    VulnConfig        // JSON: 漏洞类型/CVE/CWE/严重程度/修复方案
    FlagConfig    FlagConfig        // JSON: Flag类型/值/格式/位置
    Difficulty string; BaseScore int
    CPULimit float64; MemLimit int
    Hints    string
    Status   string                 // draft / published / archived
}
// 子类型: ImageConfig, VolumeMount, ServicePort, VulnConfig, FlagConfig, FlagLocation
// 支持导入导出: TemplateExport, TemplateImport, TemplatePreview
```

### 4.12 ServiceHealth / TargetService
```go
type ServiceHealth struct {
    ID, ServiceID uint
    Status      string            // healthy / unhealthy / unknown
    CheckedAt   time.Time
    ResponseTime int64            // 毫秒
    ErrorMsg    string
    Notified    bool
}
type TargetService struct {       // 竞技场目标
    gorm.Model
    Name string; Protocol string  // http/https/tcp
    Host string; Port int; Path string
    Enabled bool; GameID, TeamID int64
}
```

### 4.13 AdminLog
```go
type AdminLog struct {
    ID, UserID int64; Username string
    Action       string       // create/update/delete/start/pause/stop/reset/adjust_score/import
    ResourceType string       // game/team/user/score
    ResourceID   int64
    Description, IPAddress, UserAgent string
    Details      string       // JSON
    CreatedAt    time.Time
}
```

---

## 5. API 接口

### 基础信息
- 基路径: `/api/v1`
- 认证: JWT Bearer Token (Header: `Authorization: Bearer <token>`)
- WebSocket: `/ws?token=<jwt_token>`

### 5.1 认证 (`/api/v1/auth`)
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/auth/login` | 公开 | 登录 (有速率限制) |
| POST | `/auth/logout` | JWT | 登出 |
| GET  | `/auth/me` | JWT | 获取当前用户信息 |
| POST | `/auth/register` | 公开 | 注册 |
| POST | `/auth/refresh` | JWT | 刷新 Token |

### 5.2 管理员 (`/api/v1/admin`)
需要 `admin` 角色 + JWT
| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/admin/users` | 用户列表 |
| POST | `/admin/users` | 创建用户 |
| GET  | `/admin/users/:id` | 获取用户 |
| PUT  | `/admin/users/:id` | 更新用户 |
| DELETE | `/admin/users/:id` | 删除用户 |

### 5.3 队伍 (`/api/v1/teams`)
需要 JWT
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET  | `/teams/` | view_rankings | 队伍列表 |
| POST | `/teams/` | manage_teams | 创建队伍 |
| GET  | `/teams/:id` | view_rankings | 获取队伍 |
| GET  | `/teams/:id/members` | view_rankings | 队伍成员 |
| POST | `/teams/:id/members` | manage_teams | 添加成员 |
| DELETE | `/teams/:id/members/:userId` | manage_teams | 移除成员 |

### 5.4 比赛 (`/api/v1/games`)
需要 JWT
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET  | `/games/` | view_game | 比赛列表 |
| POST | `/games/` | create_game | 创建比赛 |
| GET  | `/games/:id` | view_game | 获取比赛 |
| PUT  | `/games/:id` | edit_game | 更新比赛 |
| DELETE | `/games/:id` | manage_games | 删除比赛 |
| POST | `/games/:id/start` | start_game | 启动比赛 |
| POST | `/games/:id/pause` | pause_game | 暂停比赛 |
| POST | `/games/:id/resume` | pause_game | 恢复比赛 |
| POST | `/games/:id/stop` | stop_game | 停止比赛 |
| POST | `/games/:id/reset` | manage_games | 重置比赛 |

### 5.5 题目 (`/api/v1/games/:id/challenges`)
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET  | `.../challenges` | view_game | 题目列表 |
| POST | `.../challenges` | create_challenge | 创建题目 |
| PUT  | `.../challenges/:challengeId` | edit_game | 更新题目 |
| DELETE | `.../challenges/:challengeId` | manage_games | 删除题目 |

### 5.6 Flag (`/api/v1/games/:id/flags`)
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `.../flags/submit` | submit_flag | 提交 Flag (速率限制: 100次/60秒) |
| GET  | `.../flags/history` | view_own_stats | 提交历史 |

### 5.7 容器 (`/api/v1/games/:id/containers`)
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET  | `.../containers` | view_game_stats | 容器列表 |
| POST | `.../containers/restart` | manage_infrastructure | 批量重启 |
| POST | `.../containers/:cid/restart` | manage_infrastructure | 单容器重启 |
| GET  | `.../containers/stats` | view_game_stats | 容器统计 |

### 5.8 排名
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET  | `/api/v1/games/:id/rankings` | view_rankings | 当前排名 |
| GET  | `/api/v1/games/:id/rankings/rounds/:round` | view_rankings | 指定轮次排名 |

### 5.9 轮次
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET  | `/api/v1/games/:id/rounds` | view_game | 轮次信息 |
| POST | `/api/v1/games/:id/rounds` | pause_game | 轮次控制 |

### 5.10 裁判/管理员操作 (`/api/v1/judge`)
需要 `admin` 或 `organizer` 角色
| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/judge/logs` | 管理日志 |
| POST | `/judge/games` | 创建比赛 |
| PUT  | `/judge/games/:id` | 更新比赛 |
| DELETE | `/judge/games/:id` | 删除比赛 |
| POST | `/judge/games/:id/start` | 启动 |
| POST | `/judge/games/:id/pause` | 暂停 |
| POST | `/judge/games/:id/resume` | 恢复 |
| POST | `/judge/games/:id/stop` | 停止 |
| POST | `/judge/games/:id/reset` | 重置 |
| POST | `/judge/games/:id/teams` | 添加队伍到比赛 |
| GET  | `/judge/games/:id/teams` | 获取比赛队伍 |
| DELETE | `/judge/games/:id/teams/:team_id` | 移除队伍 |
| POST | `/judge/teams` | 创建队伍 |
| PUT  | `/judge/teams/:id` | 更新队伍 |
| DELETE | `/judge/teams/:id` | 删除队伍 |
| POST | `/judge/teams/batch-import` | 批量导入队伍 |
| POST | `/judge/teams/:id/reset` | 重置队伍 |
| POST | `/judge/scores/adjust` | 人工调分 |

### 5.11 Docker 镜像 (`/api/v1/admin/images`)
需要 `admin` 角色
| 方法 | 路径 | 说明 |
|------|------|------|
| GET  | `/admin/images/` | 镜像列表 |
| GET  | `/admin/images/host/list` | 宿主机镜像 |
| POST | `/admin/images/pull` | 拉取镜像 |
| POST | `/admin/images/push` | 推送镜像 |
| POST | `/admin/images/build` | 构建镜像 |
| GET  | `/admin/images/:id` | 镜像详情 |
| GET  | `/admin/images/:id/details` | 详细信息 |
| DELETE | `/admin/images/:id` | 删除记录 |
| DELETE | `/admin/images/:id/complete` | 删除记录+宿主镜像 |
| DELETE | `/admin/images/host/:id` | 仅删宿主镜像 |
| POST | `/admin/images/:id/pull` | 拉取指定镜像 |
| POST | `/admin/images/` | 创建镜像记录 |
| PUT  | `/admin/images/:id` | 更新镜像记录 |

### 5.12 其他
| 路径 | 说明 |
|------|------|
| `/api/v1/settings` | 系统设置 (GET/PUT) |
| `/api/v1/dashboard` | 仪表盘 |
| `/api/v1/dashboard/activity` | 最近活动 |
| `/api/audit/logs` | 审计日志 (admin) |
| `/api/audit/stats` | 审计统计 (admin) |
| `/api/v1/games/:id/export/scoreboard/csv` | 导出排名 CSV |
| `/api/v1/games/:id/export/scoreboard/pdf` | 导出排名 PDF |
| `/api/v1/games/:id/export/attacks` | 导出攻击日志 |
| `/api/v1/games/:id/export/all` | 导出全部数据 |
| `/health` | 健康检查 |
| `/ws` | WebSocket 连接 |

---

## 6. 核心引擎

### CompetitionEngine (`internal/engine/engine.go`)
每个活跃比赛创建一个 `CompetitionEngine` 实例。

```
CompetitionEngine
├── game: *model.Game          // 比赛配置
├── flagSvc: *FlagService      // Flag 生成
├── gameSvc: *GameService      // 比赛服务
├── scorer: *ScoreCalculator   // 计分器
├── roundScheduler: *RoundScheduler // 轮次调度
├── flagWriter: *FlagWriter    // Docker Flag 写入器
├── dockerClient: *client.Client
├── currentRound, currentPhase, running
```

**生命周期**:
1. `NewCompetitionEngine(game)` → 初始化
2. `Start(ctx)` → 生成 Round 1 Flag → 启动 RoundScheduler goroutine
3. `Pause()` → 取消 context，设 phase=break
4. `Resume(ctx)` → 重建 RoundScheduler
5. `Stop(ctx)` → 设 phase=finished，更新数据库

**每轮流程** (在 RoundScheduler.Run 中):
1. `onRoundStart(round)` → 更新 DB → `GenerateFlags()` → `writeFlagsToContainers()`
2. 等待 `roundDuration`
3. `onRoundEnd(round)` → `CalculateRoundScores()`
4. 进入 break，等待 `breakDuration`
5. 循环或结束

### EngineManager (`internal/engine/manager.go`)
全局单例 `engine.Manager`，管理所有活跃的 `CompetitionEngine`。

```go
Manager.StartGame(game)    // 创建并启动引擎
Manager.PauseGame(gameID)  // 暂停
Manager.ResumeGame(game)   // 恢复 (重新创建引擎)
Manager.StopGame(gameID)   // 停止并删除
Manager.ShutdownAll()      // 关闭所有 (程序退出时)
```

### EngineCallbacks (`service` 包)
桥接 Service 和 Engine，避免循环依赖：
```go
service.EngineCallbacks.StartGame  = engine.Manager.StartGame
service.EngineCallbacks.PauseGame  = engine.Manager.PauseGame
service.EngineCallbacks.ResumeGame = engine.Manager.ResumeGame
service.EngineCallbacks.StopGame   = engine.Manager.StopGame
```

---

## 7. 计分系统

### 简化计分器 (`engine/scoring.go`)
`ScoreCalculator` — 基础版本：
- 统计每轮正确的 FlagSubmission
- 攻击分 = 正确提交 × attackWeight
- 防御分 = 被攻扣分 × defenseWeight
- 总分 = 攻击分 - 防御分
- Upsert RoundScore，更新 Team.Score
- UpdateRankings: 按 total_score 排序设置 rank

### 高级计分器 (`engine/scoring/zero_sum.go`)
`ZeroSumScorer` — 零和计分：

**核心规则**:
1. 所有队伍初始平分总积分池 (`initialTotal / teamCount`)
2. 攻击成功: 攻击方 +N，被攻方 -N (零和转移)
3. 首杀奖励: 第一个攻破某题的队伍额外加分
4. 防御奖励: 轮次结束未失分的队伍从防御池获得奖励
5. 总分守恒 (验证: `ValidateZeroSum()`)

```go
type ZeroSumScorer struct {
    initialTotal     float64  // 总积分池
    flagValue        float64  // 单个 Flag 分值
    firstBloodBonus  float64  // 首杀加分比例
    roundDefenseValue float64 // 每轮防御分值 (= flagValue * 0.5)
    teamScores       map[int64]*TeamScore
    firstBlood       *FirstBloodDetector
    defense          *DefenseCalculator
}
```

### DefenseCalculator (`engine/scoring/defense.go`)
- 记录每队每轮被攻破次数
- 防御奖励 = `(totalRounds - timesBreached) × defenseValue / totalRounds`
- 未被攻破的队伍从防御池分享奖励

### FirstBloodDetector (`engine/scoring/first_blood.go`)
- 跟踪每个 Flag 是否已被首次攻破
- `CheckAndRecord(flag, team, round)` → 返回是否首杀

---

## 8. 比赛状态机

### GameStateMachine (`engine/game_state_machine.go`)

**状态定义**:
```go
StatePreparing  = "preparing"   // 准备中
StateRunning    = "running"     // 比赛中
StatePaused     = "paused"      // 已暂停
StateFinished   = "finished"    // 已结束
```

**事件**: `EventStart`, `EventPause`, `EventResume`, `EventFinish`

**合法转换**:
```
preparing ──start──→ running
running   ──pause──→ paused
running   ──finish──→ finished
paused    ──resume──→ running
paused    ──finish──→ finished
finished  → (无转换)
```

**特性**:
- 回调机制: `AddCallback(func(ctx, gameID, from, to, event))`
- 数据库持久化: 可选 (`persistDB` 参数)
- 事件总线: 状态变更发布到 `game_state_<state>` 主题
- 模型同步: 自动更新 `game.Status` 和 `game.CurrentPhase`

---

## 9. 轮次管理

### RoundScheduler (`engine/round.go`)
简单版本，在 `CompetitionEngine` 内部使用：
- 定时器驱动: `roundTimer` → `onRoundEnd` → `breakTimer` → `onRoundStart`
- 事件广播: `round:start`, `round:end`, `game:finished`

### RoundManager (`engine/round_manager.go`)
高级版本，支持更多状态：
```go
type RoundPhase string  // preparation / running / break / paused / finished / scoring
```

**RoundState** 包含:
- 当前轮次、总轮数、阶段
- 轮次开始/结束时间
- 已用时间、剩余时间
- 是否暂停

**功能**:
- 暂停/恢复: 暂停时冻结计时器，恢复时补偿暂停时间
- `GetState()` → 实时状态快照
- 回调机制: `onRoundStart`, `onRoundEnd`
- 事件广播: `round:start`, `round:end`, `round:break`, `round:paused`, `round:resumed`, `game:finished`

---

## 10. Flag 系统

### FlagGenerator (`engine/flag_generator.go`)
```
flag{gameID_round_teamID_randomHex32}
```
- 16 字节随机数 = 32 位十六进制
- `GenerateBatch()` 批量生成

### FlagManager (`engine/flag.go`)
- `GenerateRoundFlags()` → 调用 FlagService
- `ValidateFlag(flag)` → SHA256 哈希比对
- 内存缓存 + 数据库回退

### FlagWriter (`engine/flag_writer.go`)
- 通过 Docker API 将 Flag 写入容器
- `WriteFlag(ctx, containerID, flagValue)`

### Flag 提交流程
1. 选手 POST `/api/v1/games/:id/flags/submit` (速率限制)
2. `FlagHandler.Submit` → `FlagService` 验证
3. 正确 → 记录 `FlagSubmission`，触发计分
4. WebSocket 推送排名更新

---

## 11. 权限系统 (RBAC)

### 角色
```go
RoleAdmin     = "admin"     // 全系统访问
RoleOrganizer = "organizer" // 比赛管理
RolePlayer    = "player"    // 参赛选手
```

### 权限
| 类别 | 权限 | Admin | Organizer | Player |
|------|------|-------|-----------|--------|
| 用户 | manage_users | ✅ | | |
| 比赛 | manage_games, create/edit/delete_game | ✅ | ✅ (except manage) | |
| 比赛 | start/pause/stop_game | ✅ | ✅ | |
| 比赛 | view_game | ✅ | ✅ | ✅ |
| 题目 | manage/create/edit/delete_challenge | ✅ | ✅ | |
| 队伍 | manage_teams | ✅ | ✅ | |
| Flag | submit_flag | ✅ | | ✅ |
| 统计 | view_game_stats | ✅ | ✅ | |
| 统计 | view_own_stats | ✅ | | ✅ |
| 排名 | view_rankings | ✅ | ✅ | ✅ |
| 基础设施 | manage_infrastructure | ✅ | | |
| 设置 | manage_settings | ✅ | | |
| 数据 | view_all_data | ✅ | | |

### 中间件链
```
JWTAuth() → RequireRole(roles...) → RequirePermission(perm)
```

---

## 12. 中间件

| 中间件 | 文件 | 说明 |
|--------|------|------|
| JWTAuth | auth.go | JWT Token 验证 |
| RequireRole | rbac_auth.go | 角色检查 |
| RequirePermission | permission.go | 权限检查 |
| LoginRateLimit | ratelimit.go | 登录速率限制 |
| FlagSubmitRateLimit | ratelimit.go | Flag 提交速率限制 |
| GlobalAPIRateLimit | ratelimit.go | 全局 API 速率限制 (100/min) |
| SecurityHeaders | security_headers.go | X-Content-Type-Options, X-Frame-Options 等 |
| CORS | cors.go | 跨域配置 |
| Logger | logger.go | 请求日志 |
| WAFMiddleware | waf_middleware.go | Web 应用防火墙 |
| Audit | audit.go | 管理操作审计日志 |
| ResourceAuth | resource_auth.go | 资源级权限 |

---

## 13. WebSocket

### WSHub (`server/ws.go`)
```go
type WSHub struct {
    clients map[*websocket.Conn]struct{}        // 所有连接
    gameSub map[string]map[*websocket.Conn]struct{} // 按游戏订阅
}
```

**功能**:
- `/ws?token=<jwt>` 连接
- JWT 认证 (query param)
- 按游戏订阅/取消订阅
- 广播消息: `ranking:update`, `round:*`, `game:*`, `game_state_*`

---

## 14. 容器管理

### Container 包 (`internal/container/`)
- `docker_client.go`: Docker 客户端初始化
- `manager.go`: 容器创建/启动/停止/重启/删除
- `limits.go`: CPU/内存资源限制
- `gorm_store.go`: 容器状态持久化到 SQLite
- `image_extensions.go`: 镜像扩展操作

### 容器生命周期
1. 比赛启动 → 为每队每题创建容器 (`TeamContainer`)
2. 容器运行中 → 健康检查 (`health_checker.go`)
3. 每轮开始 → Flag 写入容器
4. 比赛结束 → 容器清理

---

## 15. 安全 & WAF

### WAF (`internal/security/`)
- 请求规则匹配
- 攻击检测与分类
- 攻击日志记录
- 安全告警

### 全局中间件链
```
SecurityHeaders → WAF → RateLimit
```

---

## 16. 通知系统

| 通道 | 文件 | 说明 |
|------|------|------|
| 飞书 | notify/feishu.go | Webhook 通知 |
| 邮件 | notify/email.go | SMTP 邮件 |
| WebSocket | notify/websocket.go | 实时推送 |

---

## 17. 导出功能

| 格式 | 路径 | 说明 |
|------|------|------|
| CSV | `export/csv.go` | 排行榜 CSV 导出 |
| PDF | `export/pdf.go` | 排行榜 PDF 导出 (gofpdf) |
| 攻击日志 | `export_handler.go` | 攻击日志导出 |
| 全部 | `export_handler.go` | 全量数据导出 |

---

## 18. AI 分析

| 模块 | 文件 | 说明 |
|------|------|------|
| 分析器 | ai/analyzer.go | 攻击模式分析 |
| 分类器 | ai/classifier.go | 攻击类型分类 |
| 报告 | ai/report.go | 分析报告生成 |
| 规则引擎 | ai/rule_engine.go | 规则匹配 |

---

## 19. 配置

### 配置文件 (`configs/config.yaml`)
```go
type Config struct {
    Server   ServerConfig   // host, port, static_dir
    Database DatabaseConfig // sqlite_path
    // ... 其他配置
}
```

### 加载方式
```go
cfg, err := config.Load("configs/config.yaml")  // 路径可通过命令行参数覆盖
```

---

## 20. 数据库

### 当前: SQLite (GORM)
```go
database.InitDB(cfg.Database.SQLitePath)
db := database.GetDB()
```

### 自动迁移 (main.go)
```go
db.AutoMigrate(
    &model.User{}, &model.Team{}, &model.Game{},
    &model.Challenge{}, &model.TeamContainer{},
    &model.FlagRecord{}, &model.FlagSubmission{},
    &model.RoundScore{}, &model.EventLog{}, &model.AdminLog{},
)
```

### 预留的多数据源
- **PostgreSQL**: `internal/repo/postgres/` — 用户、Flag、Game、Score 仓库实现
- **Redis**: `internal/repo/redis/` — 缓存、排名
- **ClickHouse**: `internal/repo/clickhouse/` — 攻击日志

### Repo 接口
```go
type FlagRepo interface { ... }
type GameRepo interface { ... }
type ScoreRepo interface {
    SaveRoundScore(ctx, roundScore) error
    // ...
}
```

---

## 21. 前端

- 框架: React (SPA)
- 目录: `/opt/awd-arena/web/`
- 构建后由 Fiber 静态文件中间件提供
- SPA 路由: 非 `/api` 和 `/ws` 的请求都返回 `index.html`

---

## 22. 开发指南 & 约定

### 项目约定
1. **Go 模块路径**: `github.com/awd-platform/awd-arena`
2. **HTTP 框架**: Fiber v3 (`github.com/gofiber/fiber/v3`)
3. **ORM**: GORM (`gorm.io/gorm`)
4. **JSON 格式**: 结构体标签 `json:"field_name"`
5. **数据库标签**: `gorm:"primaryKey"` / `gorm:"index"` / `gorm:"default:value"`
6. **敏感字段**: `json:"-"` 不输出 (Password, Token, FlagValue)
7. **日志**: `pkg/logger` 包，`logger.Info/Error/Debug/Warn`
8. **错误处理**: Go 标准 error 返回
9. **并发安全**: `sync.Mutex` / `sync.RWMutex`
10. **上下文传递**: `context.Context` 作为第一参数

### 测试
- 计分系统测试: `internal/engine/scoring/*_test.go` (36个测试，89.1%覆盖率)
- 比赛状态机测试: `internal/engine/game_state_machine_test.go`
- 轮次管理测试: `internal/engine/round_manager_test.go`

### 添加新功能的步骤
1. 在 `internal/model/` 定义数据模型
2. 在 `internal/repo/` 定义仓库接口
3. 在 `internal/service/` 实现业务逻辑
4. 在 `internal/handler/` 添加 HTTP 处理器
5. 在 `internal/server/router.go` 注册路由
6. 在 `cmd/server/main.go` 的 AutoMigrate 中添加新模型
7. 编写测试

---

## 23. 更新日志

### 2026-03-21
- 初始项目搭建
- 完成计分系统 (零和计分 + 防御分 + 首杀检测)
- 完成比赛状态机
- 完成轮次管理器
- 完成 Flag 生成和写入系统
- 完成 RBAC 权限系统
- 完成 API 路由注册
- 完成 Docker 镜像管理
- 完成导出功能 (CSV/PDF)
- 完成 AI 分析模块骨架
- 完成通知系统骨架 (飞书/邮件/WebSocket)
- 完成安全模块 (WAF)

---

> **给 AI 助手的提示**: 修改此项目时，请将此文档同步更新。所有核心架构变更都应记录在"更新日志"中。新增的数据模型、API 接口、或核心逻辑变更都应在相应章节中补充。
