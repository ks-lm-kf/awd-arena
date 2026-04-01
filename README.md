# AWD Arena - 攻防对抗竞赛平台

<div align="center">

![AWD Arena](https://img.shields.io/badge/AWD-Arena-blue)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5+-3178C6?logo=typescript)
![Docker](https://img.shields.io/badge/Docker-20+-2496ED?logo=docker)
![License](https://img.shields.io/badge/License-MIT-green)

**一个现代化、轻量级的 AWD (Attack with Defense) 攻防对抗竞赛管理平台**

[快速开始](#快速开始) · [功能特性](#功能特性) · [部署指南](#部署指南) · [API文档](#api文档)

</div>

---

## 项目简介

AWD Arena 是一个专为网络安全竞赛设计的 AWD 攻防对抗平台。使用 Go + React 构建，提供完整的比赛生命周期管理——从比赛创建、队伍管理、Docker 容器自动编排到实时计分和排行榜 WebSocket 推送。

### 为什么选择 AWD Arena？

| 特性 | AWD Arena | 传统平台 |
|------|-----------|----------|
| 部署方式 | 单二进制 + 前端静态文件 | 多容器/复杂依赖 |
| 后端性能 | Go 高性能并发 | Python/PHP 较慢 |
| 响应式设计 | 完整支持桌面+移动端 | 部分支持 |
| 实时更新 | WebSocket 双向推送 | 轮询 |
| 安全防护 | WAF + XSS + RBAC | 基础防护 |
| 容器管理 | 自动编排每队独立靶机 | 手动管理 |

---

## 功能特性

### 🏆 核心功能

- **比赛全生命周期管理** — 创建/编辑/启动/暂停/恢复/停止/重置，完整状态机控制
- **队伍系统** — 队伍 CRUD、批量导入（CSV/TXT）、口令加入、成员管理
- **多角色权限** — admin（管理员）、organizer（裁判）、player（选手），RBAC 细粒度控制
- **Docker 容器编排** — 比赛启动自动为每队每题创建独立靶机，网络隔离，一键重启
- **Flag 自动轮转** — 每轮生成新 Flag 写入容器，支持自定义格式
- **实时计分** — 攻击分/防御分计算，零和计分算法，首杀奖励
- **WebSocket 排行榜** — 实时排名推送，无需刷新
- **选手靶机面板** — 展示 IP/端口/SSH 信息，选手自助查看

### 🔒 安全特性

- **RBAC 权限控制** — 基于角色的细粒度权限，16+ 权限项
- **WAF 防火墙** — 实时检测 SQL 注入、XSS、命令注入等攻击
- **XSS 全局防护** — 前后端双重输入验证与 HTML 转义
- **密码安全策略** — 强制大小写+数字、不能与用户名相同、首次登录强制修改
- **JWT 认证** — Token 签发/刷新，可配置过期时间
- **操作审计日志** — 所有管理操作可追溯
- **速率限制** — 登录防暴力破解、Flag 提交限频

### 🎨 用户体验

- **17 个完整页面** — 无占位符，全部功能可用
- **响应式设计** — Ant Design + Tailwind CSS，桌面移动端均适配
- **实时推送** — WebSocket 比赛状态、排行榜、容器状态即时更新
- **深色主题** — 现代化赛博风格 UI

### 📊 管理功能

- **系统设置** — 平台名称、Flag 格式、比赛参数全局配置
- **Docker 镜像管理** — 拉取/构建/推送/删除，完整镜像生命周期
- **数据导出** — 排行榜 CSV/PDF，攻击日志导出
- **审计日志** — 管理员操作完整记录
- **仪表盘** — 比赛概况、近期活动

---

## 快速开始

### 环境要求

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.22+ | 后端编译 |
| Node.js | 18+ | 前端构建 |
| Docker | 20+ | 容器编排（必需） |
| SQLite | 内置 | 默认数据库 |

### 一键部署

```bash
# 1. 克隆项目
git clone https://github.com/ks-lm-kf/awd-arena.git
cd awd-arena

# 2. 构建前端
cd web && npm install && npm run build && cd ..

# 3. 构建后端
go build -o awd-arena-server ./cmd/server/

# 4. 编辑配置
cp configs/config.yaml.example configs/config.yaml
vim configs/config.yaml

# 5. 启动
./awd-arena-server
```

### systemd 服务

```bash
# 安装为系统服务
sudo cp deployments/awd-arena.service /etc/systemd/system/
sudo systemctl enable awd-arena
sudo systemctl start awd-arena

# 查看状态
systemctl status awd-arena

# 查看日志
tail -f log/server.log
```

### 访问系统

启动后访问：**http://localhost:8080**

默认管理员：`admin` / `admin123`

> ⚠️ 首次登录后请立即在「系统设置」中修改密码！

---

## 使用指南

### 用户角色

| 角色 | 权限范围 | 说明 |
|------|----------|------|
| **admin** | 全部权限 | 系统管理员，管理用户/队伍/比赛/镜像 |
| **organizer** | 比赛管理 | 裁判，管理比赛、队伍、分数 |
| **player** | 参赛 | 选手，提交 Flag、查看靶机信息 |

### 比赛流程

```
创建比赛 → 添加题目 → 关联队伍 → 启动比赛
                                          ↓
                              自动创建容器 → Flag 轮转 → 实时计分
                                          ↓
                              暂停/恢复 → 结束比赛 → 清理容器
```

### 出题指南

1. 准备 Docker 镜像（Web/PWN/任意类型）
2. 在「Docker 镜像管理」中拉取或构建镜像
3. 创建比赛，添加题目，指定镜像名和暴露端口
4. 关联参赛队伍
5. 启动比赛，系统自动为每队创建独立靶机

---

## 技术架构

```
┌─────────────────────────────────────────────┐
│                  Frontend                    │
│    React 18 + TypeScript + Ant Design       │
│    Vite + TanStack Query + Zustand          │
└──────────────────┬──────────────────────────┘
                   │ HTTP / WebSocket
┌──────────────────▼──────────────────────────┐
│               Go Backend                     │
│    Fiber v3 + GORM + SQLite                  │
│    JWT Auth + RBAC + WAF                     │
├──────────────────────────────────────────────┤
│  Engine: 状态机 / Flag轮转 / 计分引擎        │
│  Container: Docker API / 网络隔离            │
│  EventBus: WebSocket 广播                    │
└──────────────────┬──────────────────────────┘
                   │ Docker API
┌──────────────────▼──────────────────────────┐
│              Docker Engine                   │
│    每队独立容器 / 网络隔离 / 资源限制         │
└──────────────────────────────────────────────┘
```

---

## API 文档

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/register` | 注册 |
| PUT | `/api/v1/auth/change-password` | 修改密码 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| GET | `/api/v1/auth/me` | 当前用户信息 |

### 比赛

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/games` | 比赛列表 |
| POST | `/api/v1/games` | 创建比赛 |
| GET | `/api/v1/games/:id` | 比赛详情 |
| PUT | `/api/v1/games/:id` | 更新比赛 |
| POST | `/api/v1/games/:id/start` | 启动 |
| POST | `/api/v1/games/:id/pause` | 暂停 |
| POST | `/api/v1/games/:id/resume` | 恢复 |
| POST | `/api/v1/games/:id/stop` | 停止 |
| POST | `/api/v1/games/:id/reset` | 重置 |

### Flag & 排行榜

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/games/:id/flags/submit` | 提交 Flag |
| GET | `/api/v1/games/:id/rankings` | 排行榜 |
| GET | `/api/v1/games/:id/my-containers` | 我的靶机 |

### 管理后台

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST/PUT/DELETE | `/api/v1/admin/users` | 用户管理 |
| GET/PUT | `/api/v1/teams` | 队伍管理 |
| GET/PUT | `/api/v1/settings` | 系统设置 |
| GET/POST | `/api/v1/admin/images/*` | Docker 镜像管理 |
| GET | `/api/v1/dashboard` | 仪表盘 |

---

## 部署指南

### 生产环境建议

```yaml
# configs/config.yaml
server:
  host: 0.0.0.0
  port: 8080
  jwt_secret: "your-strong-random-secret-here"  # 必须修改！
  static_dir: "web/dist"

security:
  jwt_expire_hours: 24
  rate_limit: 100
  flag_submit_limit: 100
  blocklist_enabled: true
```

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name awd.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 安全建议

- ✅ 修改默认密码和 JWT Secret
- ✅ 启用 HTTPS（Nginx/Caddy 配置 SSL）
- ✅ 限制管理后台访问 IP
- ✅ 定期备份 `data/awd.db`
- ✅ 使用防火墙限制 Docker 网络访问

---

## 项目结构

```
awd-arena/
├── cmd/server/          # 入口
├── internal/
│   ├── config/          # 配置
│   ├── database/        # 数据库
│   ├── handler/         # HTTP 处理器
│   ├── middleware/       # 中间件（Auth/RBAC/WAF/CORS）
│   ├── model/           # 数据模型
│   ├── service/         # 业务逻辑
│   ├── engine/          # 比赛引擎（状态机/Flag/计分）
│   ├── container/       # Docker 容器管理
│   ├── eventbus/        # 事件总线
│   ├── server/          # 路由 & WebSocket
│   ├── security/        # WAF & 安全
│   └── monitor/         # 健康监控
├── web/                 # React 前端
│   └── src/
│       ├── api/         # API 客户端
│       ├── pages/       # 页面组件（17个）
│       ├── stores/      # 状态管理
│       └── hooks/       # 自定义 Hooks
├── configs/             # 配置文件
├── deployments/         # 部署文件
└── data/                # SQLite 数据库
```

---

## 已部署的 CTF 题目镜像

| 镜像名称 | 类型 | 说明 |
|----------|------|------|
| `awd-web-ezsql` | Web | SQL 注入题 |
| `awd-pwn-ezstack` | PWN | 栈溢出题 |

---

## 常见问题

**Q: 忘记管理员密码？**
```bash
sqlite3 data/awd.db "UPDATE users SET password='$(go run pkg/crypto/main.go admin123)' WHERE username='admin';"
```

**Q: 容器无法启动？**
检查 Docker 服务状态：`systemctl status docker`，确认镜像已拉取

**Q: WebSocket 连接失败？**
检查 Nginx 代理配置是否包含 Upgrade 头

**Q: 排行榜不更新？**
确保 WebSocket 连接正常，检查浏览器控制台

---

## 许可证

[MIT License](LICENSE)

---

<div align="center">
Made with ❤️ by AWD Arena Team
</div>
