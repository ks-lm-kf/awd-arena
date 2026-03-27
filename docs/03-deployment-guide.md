# AWD Arena — 部署使用指南

> **版本**: v1.0  
> **日期**: 2026-03-19  
> **适用平台**: Linux (推荐) / Windows

---

## 目录

1. [环境要求](#1-环境要求)
2. [快速开始 — Linux](#2-快速开始--linux)
3. [快速开始 — Windows](#3-快速开始--windows)
4. [手动构建](#4-手动构建)
5. [配置说明](#5-配置说明)
6. [Docker 部署](#6-docker-部署)
7. [前端开发](#7-前端开发)
8. [常见问题 (FAQ)](#8-常见问题-faq)
9. [目录结构说明](#9-目录结构说明)

---

## 1. 环境要求

### 1.1 硬件要求

| 指标 | 最低要求 | 推荐配置 |
|------|---------|---------|
| CPU | 4 核 | 8 核+ |
| 内存 | 8 GB | 16 GB+ |
| 磁盘 | 50 GB SSD | 100 GB+ SSD |
| 网络 | 百兆 | 千兆 |

> 💡 内存和磁盘需求取决于参赛队伍数量和靶机容器数量。50+ 队伍建议 32GB RAM。

### 1.2 软件依赖

| 依赖 | 版本要求 | 用途 |
|------|---------|------|
| **Go** | 1.21+ (推荐 1.24+) | 编译后端 |
| **Node.js** | 18+ | 前端开发构建 |
| **Docker** | 20.10+ | 靶机容器运行时 |
| **Docker Compose** | v2+ | 基础设施编排 |
| **Git** | 任意 | 源码获取 |
| **Make** | 任意 (Linux) | 构建工具 |

> 💡 数据库使用 SQLite，无需额外安装 PostgreSQL 等数据库服务。

### 1.3 前置条件检查

**Linux:**

```bash
# 检查各依赖版本
go version          # go version go1.24.x linux/amd64
node --version      # v18.x.x+
docker --version    # Docker version 20.10.x+
docker compose version  # Docker Compose version v2.x.x+
make --version      # GNU Make 4.x+
```

**Windows:**

```powershell
go version          # go version go1.24.x windows/amd64
node --version      # v18.x.x+
docker --version    # Docker version 20.10.x+
docker compose version
```

---

## 2. 快速开始 — Linux

### 2.1 自动安装 Docker（可选）

如果尚未安装 Docker，可运行安装脚本：

```bash
chmod +x scripts/install.sh
sudo ./scripts/install.sh
```

该脚本支持 Ubuntu/Debian 和 CentOS/RHEL，会自动安装 Docker CE 和 Docker Compose 插件。

### 2.2 一键部署

```bash
# 1. 克隆项目
git clone <repository-url> awd-arena
cd awd-arena

# 2. 构建
chmod +x scripts/build.sh
./scripts/build.sh

# 3. 部署（自动编译 + 启动服务）
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

部署脚本会自动完成：
- ✅ 检查端口 8080 是否可用
- ✅ 创建数据目录
- ✅ 编译后端二进制
- ✅ 启动 AWD Arena 服务（数据库自动初始化，无需手动建表）

### 2.3 安装二进制 + Systemd 服务

如果需要将 AWD Arena 安装为系统服务：

```bash
# 安装二进制到 /usr/local/bin
sudo ./scripts/install.sh

# 同时创建 Systemd 服务
sudo ./scripts/install.sh --systemd
```

创建后的服务管理命令：

```bash
sudo systemctl start awd-arena     # 启动
sudo systemctl stop awd-arena      # 停止
sudo systemctl restart awd-arena   # 重启
sudo systemctl status awd-arena    # 查看状态
sudo journalctl -u awd-arena -f    # 查看日志
```

### 2.4 访问平台

部署成功后：

| 项目 | 地址 |
|------|------|
| 管理后台 | http://localhost:8080 |
| API 文档 | http://localhost:8080/swagger/ （如已启用） |
| 默认账号 | **admin** / **admin123** |

> 🔐 首次登录后请立即修改默认密码！

---

## 3. 快速开始 — Windows

### 3.1 前置条件

1. 安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/) 并确保 Docker Engine 运行中（仅 Docker 部署方式需要）
2. 安装 [Go](https://go.dev/dl/) 1.21+
3. 安装 [Node.js](https://nodejs.org/) 18+（前端构建需要）
4. 确保已启用 WSL2（推荐，以获得完整的网络隔离支持）

### 3.2 一键部署（推荐）

使用 Windows 部署脚本，自动编译并启动服务：

```powershell
# 1. 克隆项目
git clone <repository-url> awd-arena
cd awd-arena

# 2. 一键部署
powershell -ExecutionPolicy Bypass -File scripts/deploy.ps1
```

`deploy.ps1` 会自动完成：
- ✅ 检查 Go 环境
- ✅ 检查端口 8080 是否可用
- ✅ 创建数据目录
- ✅ 编译 `awd-arena.exe` 和 `awd-cli.exe`
- ✅ 检查前端是否已构建
- ✅ 启动 AWD Arena 服务

> 💡 数据库使用 SQLite（`data/awd.db`），无需单独安装数据库服务。首次启动时自动创建表。

### 3.3 使用 Docker Compose 部署

```powershell
# 1. 克隆项目
git clone <repository-url> awd-arena
cd awd-arena

# 2. 启动所有服务
docker compose -f deployments/docker-compose.yml up -d
```

Docker Compose 会自动启动 Redis、NATS 和 AWD Server，无需手动迁移数据库。

### 3.4 手动编译部署

```powershell
# 1. 编译 Windows 二进制
cd awd-arena
go build -o build\awd-arena.exe .\cmd\server
go build -o build\awd-cli.exe .\cmd\cli

# 2. 启动服务（数据库自动初始化）
.\build\awd-arena.exe
```

### 3.5 多平台构建

使用 Windows 构建脚本编译 Linux + Windows 全平台二进制：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build.ps1

# 可通过环境变量指定版本号
$env:VERSION="1.0.0"; powershell -ExecutionPolicy Bypass -File scripts/build.ps1
```

构建产物输出到 `dist/` 目录，包含 `.zip` 压缩包。

### 3.6 访问平台

| 项目 | 地址 |
|------|------|
| 管理后台 | http://localhost:8080 |
| 默认账号 | **admin** / **admin123** |

---

## 4. 手动构建

### 4.1 使用 Makefile（Linux / macOS）

```bash
# 构建所有平台的全部二进制
make build-all

# 仅构建 Linux 服务器
make build-linux

# 仅构建 Windows 服务器
make build-windows

# 仅构建 CLI 工具
make build-cli-linux
make build-cli-windows

# 构建前端
make build-frontend

# 清理构建产物
make clean
```

构建产物输出到 `build/` 目录。

### 4.2 使用 build.sh（Linux）

```bash
chmod +x scripts/build.sh
./scripts/build.sh
```

该脚本会自动：
1. 检测 Go 环境
2. 编译 Linux + Windows 全部二进制（交叉编译）
3. 打包为 `dist/awd-arena-{version}-{platform}.tar.gz` / `.zip`

可通过环境变量指定版本号：

```bash
VERSION=1.0.0 ./scripts/build.sh
```

### 4.3 使用 build.ps1（Windows）

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
```

该脚本会自动：
1. 编译 Linux amd64 + Windows amd64 全部二进制
2. 尝试构建前端（`web/` 目录）
3. 打包为 `dist/awd-arena-{version}-windows-amd64.zip`

### 4.4 构建产物说明

| 文件 | 平台 | 说明 |
|------|------|------|
| `awd-arena` / `awd-arena.exe` | Linux / Windows | 主服务二进制 |
| `awd-cli` / `awd-cli.exe` | Linux / Windows | 命令行管理工具 |
| `awd-migrator` / `awd-migrator.exe` | Linux / Windows | 数据库迁移工具（可选） |

### 4.5 运行主服务

```bash
# Linux
./build/awd-arena server --config configs/config.yaml

# Windows
.\build\awd-arena.exe server --config configs\config.yaml

# 使用 CLI 工具
./build/awd-cli --help
```

---

## 5. 配置说明

### 5.1 配置文件

主配置文件为 `configs/config.yaml`，完整示例：

```yaml
server:
  host: 0.0.0.0          # 监听地址，0.0.0.0 表示所有网卡
  port: 8080              # HTTP 端口
  tls_port: 8443          # HTTPS 端口（可选）
  jwt_secret: "${JWT_SECRET}"  # JWT 密钥，建议通过环境变量设置
  static_dir: "web/dist"  # 前端静态文件目录（默认值，可省略）

database:
  sqlite_path: "data/awd.db"  # SQLite 数据库文件路径（默认值，可省略）

nats:
  url: "nats://localhost:4222"

docker:
  host: "unix:///var/run/docker.sock"   # Linux
  # host: "npipe:////./pipe/docker_engine"  # Windows

security:
  jwt_expire_hours: 24       # JWT 过期时间（小时）
  rate_limit: 100            # 全局限流（请求/分钟）
  flag_submit_limit: 100     # Flag 提交限流（次/分钟/队伍）
  blocklist_enabled: true    # 是否启用 IP 黑名单

game:
  default_round_duration: 5m   # 默认轮次时长
  default_break_duration: 2m   # 默认休息时长
  max_containers_per_team: 5   # 每队最大容器数

ai:
  enabled: true                      # 是否启用 AI 分析
  rule_engine_path: "./rules/"       # 规则引擎路径
  onnx_model_path: "./models/"       # ONNX 模型路径

log_level: "info"            # 日志级别: debug / info / warn / error
```

### 5.2 配置字段详解

#### server（服务配置）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `host` | string | `0.0.0.0` | 服务监听地址 |
| `port` | int | `8080` | HTTP 服务端口 |
| `tls_port` | int | `8443` | HTTPS 端口，配置证书后生效 |
| `jwt_secret` | string | 环境变量 | JWT 签名密钥，**生产环境务必修改** |
| `static_dir` | string | `web/dist` | 前端静态文件目录 |

#### database（数据库配置）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `sqlite_path` | string | `data/awd.db` | SQLite 数据库文件路径 |

> 💡 数据库使用 SQLite，首次启动时通过 GORM AutoMigrate 自动创建所有表，无需手动执行迁移。

#### security（安全配置）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `jwt_expire_hours` | int | `24` | Token 过期时间 |
| `rate_limit` | int | `100` | API 全局限流 |
| `flag_submit_limit` | int | `100` | Flag 提交频率限制 |
| `blocklist_enabled` | bool | `true` | IP 黑名单功能开关 |

#### game（竞赛配置）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `default_round_duration` | duration | `5m` | 默认轮次时长 |
| `default_break_duration` | duration | `2m` | 默认轮间休息时长 |
| `max_containers_per_team` | int | `5` | 每队最大靶机容器数 |

### 5.3 环境变量覆盖

配置文件中支持 `${ENV_VAR}` 语法引用环境变量：

```bash
# 设置 JWT 密钥（必须）
export JWT_SECRET="your-secret-key-here"

# 覆盖 NATS 连接地址（可选）
export NATS_URL="nats://nats-host:4222"

# 启动服务
./awd-arena server --config configs/config.yaml
```

---

## 6. Docker 部署

### 6.1 使用 Docker Compose（推荐生产方式）

```bash
# 启动所有服务（Redis + NATS + AWD Server）
docker compose -f deployments/docker-compose.yml up -d

# 查看服务状态
docker compose -f deployments/docker-compose.yml ps

# 查看日志
docker compose -f deployments/docker-compose.yml logs -f awd-server

# 停止所有服务
docker compose -f deployments/docker-compose.yml down

# 停止并清除数据卷（⚠️ 会删除所有数据）
docker compose -f deployments/docker-compose.yml down -v
```

### 6.2 服务组件

Docker Compose 包含以下服务：

| 服务 | 镜像 | 端口 | 说明 |
|------|------|------|------|
| `awd-server` | 本地构建 | 8080 | AWD Arena 主服务 |
| `redis` | redis:8-alpine | 6379 | 缓存 / 排行榜 |
| `nats` | nats:alpine | 4222, 8222 | 消息队列 / 事件总线（JetStream） |

> 💡 数据库使用 SQLite，数据文件存储在 `./data/awd.db`，通过 Volume 挂载到容器内。

### 6.3 数据持久化

| 路径 / Volume | 说明 |
|---------------|------|
| `./data` (宿主机) → `/data` (容器) | SQLite 数据库文件 + 其他数据 |
| `redisdata` (Volume) | Redis 缓存数据 |
| `natsdata` (Volume) | NATS JetStream 消息持久化 |

### 6.4 自定义 Docker 构建

如需修改 Dockerfile 后重新构建：

```bash
# 重新构建 AWD Server 镜像
docker compose -f deployments/docker-compose.yml build awd-server

# 重新构建并启动
docker compose -f deployments/docker-compose.yml up -d --build awd-server
```

Dockerfile 采用多阶段构建：
- **构建阶段**: `golang:1.24-alpine`，编译 Go 二进制
- **运行阶段**: `alpine:3.20`，仅包含二进制 + 证书 + Docker CLI

---

## 7. 前端开发

### 7.1 技术栈

- **框架**: React 19 + TypeScript 5
- **构建工具**: Vite 6
- **UI 库**: Ant Design 5
- **状态管理**: Zustand
- **图表**: Recharts
- **HTTP 客户端**: Axios

### 7.2 目录结构

前端源码位于 `web/` 目录。构建后的静态文件输出到 `web/dist/`，后端启动时自动挂载提供访问。

### 7.3 本地开发

```bash
cd web

# 安装依赖
npm install
# 或
yarn install

# 启动开发服务器
npm run dev
# 或
yarn dev
```

开发服务器默认运行在 `http://localhost:5173`，支持热模块替换（HMR）。

### 7.4 构建

```bash
# 生产构建
npm run build

# 构建产物输出到 web/dist/ 目录
# 后端启动时会自动加载该目录的静态文件，无需额外配置
```

也可以通过 Makefile 构建：

```bash
make build-frontend
```

---

## 8. 常见问题 (FAQ)

### Q: 部署脚本报 "Port 8080 is already in use"

端口被占用。检查并释放端口：

```bash
# Linux — 查看占用进程
ss -tlnp | grep :8080
lsof -i :8080

# 或修改 configs/config.yaml 中的 server.port
```

### Q: 数据库文件在哪里？

默认位于 `data/awd.db`（相对于项目根目录）。可在 `configs/config.yaml` 中通过 `database.sqlite_path` 修改路径。首次启动时自动创建，无需手动初始化。

### Q: Docker 权限不足 (permission denied)

```bash
# 将当前用户加入 docker 组
sudo usermod -aG docker $USER
newgrp docker
```

### Q: Windows 下网络隔离功能不可用

Windows Docker Desktop 的网络能力有限。建议使用 WSL2 以获得完整的网络隔离支持（VLAN、iptables 等）。

### Q: 如何修改默认密码？

首次登录后在管理后台修改，或通过 CLI：

```bash
./build/awd-cli user update-password --username admin --new-password "your-new-password"
```

### Q: 如何重置数据库？

```bash
# 删除 SQLite 数据库文件
rm data/awd.db

# 重新启动服务（会自动重建数据库）
./build/awd-arena server --config configs/config.yaml
```

Docker 部署时：

```bash
# 停止服务并清除数据
docker compose -f deployments/docker-compose.yml down -v
rm -rf ./data

# 重新启动（会自动初始化）
docker compose -f deployments/docker-compose.yml up -d
```

### Q: 如何查看服务日志？

```bash
# Docker 方式
docker compose -f deployments/docker-compose.yml logs -f awd-server

# 二进制方式
# 日志输出到 stdout，可用 systemd 或 nohup 管理
journalctl -u awd-arena -f   # systemd 模式
```

### Q: 比赛中靶机容器异常怎么办？

1. 检查 Docker Engine 状态：`docker info`
2. 查看容器日志：在管理后台的「容器管理」页面
3. 手动重启容器：`docker restart <container_id>`
4. 检查宿主机资源：`docker stats`

---

## 9. 目录结构说明

```
awd-arena/
├── cmd/                          # 程序入口
│   ├── server/                   # 主服务 (awd-arena)
│   │   └── main.go
│   ├── cli/                      # 命令行工具 (awd-cli)
│   │   └── main.go
│   └── migrator/                 # 数据库迁移工具 (awd-migrator)
│       └── main.go
│
├── internal/                     # 内部包
│   ├── config/                   # 配置加载
│   ├── middleware/                # HTTP 中间件 (JWT/限流/日志/CORS)
│   ├── server/                   # HTTP 服务器 & 路由 & WebSocket
│   ├── engine/                   # 竞赛引擎 (Round/评分/Flag/模式)
│   │   └── mode/                 # 竞赛模式实现
│   ├── container/                # Docker 容器管理
│   ├── network/                  # 网络管理 (隔离/流量镜像/PCAP)
│   ├── security/                 # 安全层 (WAF/IDS/告警)
│   │   └── rules/                # WAF 规则文件
│   ├── ai/                       # AI 分析层
│   ├── monitor/                  # 监控 & 指标 & WebSocket 推送
│   ├── model/                    # 数据模型
│   ├── repo/                     # 数据访问层
│   ├── service/                  # 业务逻辑层
│   ├── handler/                  # HTTP Handler
│   └── eventbus/                 # 事件总线 (NATS)
│
├── pkg/                          # 公共工具包
│   ├── logger/                   # 日志封装
│   ├── validator/                # 参数校验
│   ├── crypto/                   # 加密工具
│   └── httputil/                 # HTTP 工具
│
├── api/                          # API 定义
│   └── openapi.yaml              # OpenAPI 3.1 规范
│
├── migrations/                   # 数据库迁移 SQL（可选，GORM AutoMigrate 为主）
│
├── configs/                      # 配置文件
│   └── config.yaml               # 主配置文件
│
├── web/                          # 前端源码
│   ├── src/                      # React + TypeScript 源码
│   ├── dist/                     # 构建产物（后端自动挂载）
│   └── package.json
│
├── data/                         # 数据目录（SQLite 数据库文件等）
│
├── scripts/                      # 工具脚本
│   ├── install.sh                # Linux 安装脚本 (Docker + Systemd)
│   ├── deploy.sh                 # Linux 一键部署脚本
│   ├── deploy.ps1                # Windows 一键部署脚本
│   ├── build.sh                  # Linux 构建 + 打包脚本
│   └── build.ps1                 # Windows 多平台构建脚本
│
├── deployments/                  # 部署配置
│   ├── docker/
│   │   └── Dockerfile            # Docker 镜像构建
│   └── docker-compose.yml        # Docker Compose 编排 (Redis + NATS + Server)
│
├── build/                        # 构建产物 (gitignore)
├── dist/                         # 分发包 (gitignore)
│
├── docs/                         # 项目文档
│   ├── 01-requirements.md        # 需求分析文档
│   ├── 02-architecture.md        # 架构设计文档
│   └── 03-deployment-guide.md    # 部署指南（本文档）
│
├── tests/                        # 测试
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── Makefile                      # Make 构建规则
├── go.mod                        # Go 模块定义
├── go.sum                        # Go 依赖校验
└── README.md
```

---

> **相关文档**:  
> - [需求分析文档](./01-requirements.md)  
> - [架构设计文档](./02-architecture.md)
