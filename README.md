# AWD Arena - 攻防对抗竞赛平台

<div align="center">

![AWD Arena](https://img.shields.io/badge/AWD-Arena-blue)
![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-5+-3178C6?logo=typescript)
![Docker](https://img.shields.io/badge/Docker-24+-2496ED?logo=docker)
![License](https://img.shields.io/badge/License-MIT-green)

**一个现代化、轻量级的 AWD (Attack with Defense) 攻防对抗竞赛管理平台**

[快速开始](#快速开始) · [功能特性](#功能特性) · [部署指南](#部署指南) · [API 文档](#api-文档) · [项目结构](#项目结构)

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
- **队伍系统** — 队伍 CRUD、JSON 批量导入、口令加入、成员管理
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
- **密码安全策略** — 强制大小写+数字+特殊字符、不能与用户名相同、首次登录强制修改
- **JWT 认证** — Token 签发/刷新，可配置过期时间
- **操作审计日志** — 所有管理操作可追溯
- **速率限制** — 登录防暴力破解、Flag 提交限频
- **错误信息脱敏** — 后端错误响应不泄露内部细节（文件路径、Docker API 错误等）
- **网络隔离** — iptables 自动隔离各队容器网络，确保跨队流量阻断

### 🎨 用户体验

- **17 个完整页面** — 无占位符，全部功能可用
- **响应式设计** — Ant Design + Tailwind CSS，桌面移动端均适配
- **实时推送** — WebSocket 比赛状态、排行榜、容器状态即时更新
- **深色主题** — 现代化赛博风格 UI

### 📊 管理功能

- **系统设置** — 平台名称、Flag 格式、比赛参数全局配置
- **Docker 镜像管理** — 拉取/构建/推送/删除，完整镜像生命周期
- **数据导出** — 排行榜 CSV/HTML、攻击日志导出
- **审计日志** — 管理员操作完整记录
- **仪表盘** — 比赛概况、近期活动

---

## 快速开始

### 环境要求

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.24+ | 后端编译 |
| Node.js | 18+ | 前端构建 |
| Docker | 24+ | 容器编排（必需） |
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

默认管理员：`admin` / `Admin@2026`

> ⚠️ 首次登录将被强制修改密码，请妥善保管新密码。

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
| POST | `/api/v1/auth/logout` | 登出 |
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| GET | `/api/v1/auth/me` | 当前用户信息 |
| PUT | `/api/v1/auth/change-password` | 修改密码 |
| POST | `/api/v1/auth/password` | 修改密码（备用） |

### 比赛

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/games` | 比赛列表 |
| POST | `/api/v1/games` | 创建比赛 |
| GET | `/api/v1/games/:id` | 比赛详情 |
| PUT | `/api/v1/games/:id` | 更新比赛 |
| DELETE | `/api/v1/games/:id` | 删除比赛 |
| POST | `/api/v1/games/:id/start` | 启动 |
| POST | `/api/v1/games/:id/pause` | 暂停 |
| POST | `/api/v1/games/:id/resume` | 恢复 |
| POST | `/api/v1/games/:id/stop` | 停止 |
| POST | `/api/v1/games/:id/reset` | 重置 |

> **Note**: `round_duration` and `break_duration` are integer values in **seconds** (e.g., use `300` for 5 minutes, not `"5m"`).

### 题目

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/challenges` | 题目列表 |
| POST | `/api/v1/games/:id/challenges` | 创建题目 |
| PUT | `/api/v1/games/:id/challenges/:challengeId` | 更新题目 |
| DELETE | `/api/v1/games/:id/challenges/:challengeId` | 删除题目 |

### Flag & 排行榜

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/games/:id/flags/submit` | 提交 Flag |
| GET | `/api/v1/games/:id/flags/history` | 提交历史 |
| GET | `/api/v1/games/:id/rankings` | 排行榜 |
| GET | `/api/v1/games/:id/rankings/rounds/:round` | 单轮排行 |

### 容器

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/containers` | 容器列表 |
| GET | `/api/v1/games/:id/containers/stats` | 容器统计 |
| POST | `/api/v1/games/:id/containers/restart` | 批量重启 |
| POST | `/api/v1/games/:id/containers/:cid/restart` | 重启单个 |
| GET | `/api/v1/games/:id/my-containers` | 我的靶机 |
| GET | `/api/v1/games/:id/my-machines` | 我的靶机（别名） |

### 轮次

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/rounds` | 轮次列表 |
| POST | `/api/v1/games/:id/rounds` | 轮次控制 |

### 安全 & 攻击日志

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/alerts` | 比赛安全告警 |
| GET | `/api/v1/games/:id/attacks` | 比赛攻击日志 |
| GET | `/api/v1/security/waf/rules` | WAF 规则列表 |
| GET | `/api/v1/security/attacks` | 全局攻击日志 |
| GET | `/api/v1/waf/rules` | WAF 规则（别名） |

### 队伍

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/teams` | 队伍列表 |
| POST | `/api/v1/teams` | 创建队伍 |
| GET | `/api/v1/teams/:id` | 队伍详情 |
| GET | `/api/v1/teams/:id/members` | 队伍成员 |
| POST | `/api/v1/teams/:id/members` | 添加成员 |
| DELETE | `/api/v1/teams/:id/members/:userId` | 移除成员 |
| DELETE | `/api/v1/teams/:id` | 删除队伍 |

### 管理后台

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/users` | 用户列表 |
| POST | `/api/v1/admin/users` | 创建用户 |
| GET | `/api/v1/admin/users/:id` | 用户详情 |
| PUT | `/api/v1/admin/users/:id` | 更新用户 |
| DELETE | `/api/v1/admin/users/:id` | 删除用户 |

### Docker 镜像

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/docker-images` | 镜像列表 |
| GET | `/api/v1/docker-images/host/list` | 宿主机镜像 |
| GET | `/api/v1/docker-images/:id` | 镜像详情 |
| POST | `/api/v1/docker-images` | 创建镜像记录 |
| PUT | `/api/v1/docker-images/:id` | 更新镜像记录 |
| DELETE | `/api/v1/docker-images/:id` | 删除镜像记录 |
| POST | `/api/v1/docker-images/:id/pull` | 拉取镜像 |
| GET | `/api/v1/admin/images` | 管理镜像列表 |
| POST | `/api/v1/admin/images/pull` | 拉取镜像（管理） |
| POST | `/api/v1/admin/images/push` | 推送镜像 |
| POST | `/api/v1/admin/images/build` | 构建镜像 |
| GET | `/api/v1/admin/images/:id/details` | 镜像详细信息 |
| DELETE | `/api/v1/admin/images/:id/complete` | 完整删除（DB+主机） |
| DELETE | `/api/v1/admin/images/host/:id` | 从主机删除 |

### 裁判操作

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/judge/logs` | 管理操作日志 |
| POST | `/api/v1/judge/games` | 创建比赛 |
| PUT | `/api/v1/judge/games/:id` | 更新比赛 |
| DELETE | `/api/v1/judge/games/:id` | 删除比赛 |
| POST | `/api/v1/judge/games/:id/start` | 启动比赛 |
| POST | `/api/v1/judge/games/:id/pause` | 暂停比赛 |
| POST | `/api/v1/judge/games/:id/resume` | 恢复比赛 |
| POST | `/api/v1/judge/games/:id/stop` | 停止比赛 |
| POST | `/api/v1/judge/games/:id/reset` | 重置比赛 |
| POST | `/api/v1/judge/games/:id/teams` | 关联队伍 |
| GET | `/api/v1/judge/games/:id/teams` | 获取比赛队伍 |
| DELETE | `/api/v1/judge/games/:id/teams/:team_id` | 移除队伍 |
| POST | `/api/v1/judge/teams` | 创建队伍 |
| PUT | `/api/v1/judge/teams/:id` | 更新队伍 |
| DELETE | `/api/v1/judge/teams/:id` | 删除队伍 |
| POST | `/api/v1/judge/teams/batch-import` | JSON 批量导入 |
| POST | `/api/v1/judge/teams/:id/reset` | 重置队伍 |
| POST | `/api/v1/judge/scores/adjust` | 分数调整 |

### 数据导出

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/games/:id/export/ranking/csv` | 排行榜 CSV |
| GET | `/api/v1/games/:id/export/ranking/pdf` | 排行榜 HTML |
| GET | `/api/v1/games/:id/export/attacks` | 攻击日志导出 |
| GET | `/api/v1/games/:id/export/all` | 全量导出 |

### 审计

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/audit/logs` | 审计日志 |
| GET | `/api/v1/audit/stats` | 审计统计 |

### 仪表盘 & 设置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/dashboard` | 仪表盘概况 |
| GET | `/api/v1/dashboard/activity` | 近期活动 |
| GET | `/api/v1/settings` | 系统设置 |
| PUT | `/api/v1/settings` | 更新设置 |

### WebSocket

| 路径 | 说明 |
|------|------|
| `/ws` | WebSocket 连接（Token 通过查询参数验证） |

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
│   ├── network/         # 网络隔离（iptables）
│   ├── security/        # WAF & 安全
│   └── server/          # 路由 & WebSocket
├── pkg/                  # 公共工具包（crypto, logger）
├── web/                 # React 前端
│   └── src/
│       ├── api/         # API 客户端
│       ├── pages/       # 页面组件（17个）
│       ├── stores/      # 状态管理
│       └── hooks/       # 自定义 Hooks
├── configs/             # 配置文件
├── deployments/         # 部署文件
├── scripts/             # 构建 & 部署脚本
├── migrations/          # SQL 迁移脚本
└── challenges/          # CTF 题目示例
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
sqlite3 data/awd.db "UPDATE users SET password='$(go run pkg/crypto/main.go NewPassword123)' WHERE username='admin';"
```

**Q: 容器无法启动？**
检查 Docker 服务状态：`systemctl status docker`，确认镜像已拉取

**Q: WebSocket 连接失败？**
检查 Nginx 代理配置是否包含 Upgrade 头

**Q: 排行榜不更新？**
确保 WebSocket 连接正常，检查浏览器控制台

---

## 更新日志

### 2026-04 — 比赛状态机修复

**Bug 修复：**
- 修复 `mapModelToGameState` 未处理 `status="active"` 导致运行中/暂停的比赛恢复后状态回退为 preparing
- 修复 `updateModelState` 将 Running/Paused 状态错误映射为 `status="running"`/`"paused"`，应为 `"active"` + `currentPhase` 区分

### 2026-04 — 全面安全审计 & Bug 修复

修复 GitHub Issues #36–#82，共计 50+ 项问题：

**安全修复：**
- 网络子网计算不一致导致 iptables 规则失效（#66）
- Handler 错误响应泄露内部信息（#78）
- JWT Refresh Token 泄漏容忍（#57）
- 速率限制首个请求未记录（#64）
- 密码强度要求增强：必须包含特殊字符（#63）

**稳定性修复：**
- EventBus Close() 后 Publish() 触发 nil map panic（#67）
- Engine Pause/Resume 状态未持久化到数据库（#74）
- 事件重复广播（round:start, round:end, game:finished）（#75）
- 休息阶段计时器暂停后继续倒计时（#76）
- Scoring 模块 8 处数据库写入错误静默忽略（#77）
- Flag 生成密码学错误被丢弃（#68）
- Context 未传播到数据库操作（#79）

**前端修复：**
- 比赛管理页面导航到不存在的路由（#69）
- 修改密码后前端状态未清理导致幽灵登录（#70）
- 前端调用不存在的后端 API（resetPassword, toggleStatus, containerDetail）（#73）
- Dashboard 排行榜引用不存在的字段（#72）
- DockerImages 表单 Select 组件值不更新（#80）
- 侧边栏移动端自适应问题（#81）
- 大量未使用导入和死代码清理（#82）

---

## 许可证

[MIT License](LICENSE)

---

<div align="center">
Made with ❤️ by AWD Arena Team
</div>
