# AWD Arena - 攻防对抗竞赛平台

<div align="center">

![AWD Arena](https://img.shields.io/badge/AWD-Arena-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)
![License](https://img.shields.io/badge/License-MIT-green)

**一个现代化、轻量级的 AWD (Attack with Defense) 攻防对抗竞赛管理平台**

</div>

---

## 目录

- [项目简介](#项目简介)
- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [使用指南](#使用指南)
- [API文档](#api文档)
- [部署指南](#部署指南)
- [常见问题](#常见问题)

---

## 项目简介

AWD Arena 是一个专为网络安全竞赛设计的 AWD 攻防对抗平台。提供完整的比赛生命周期管理，从比赛创建、队伍管理到实时计分和排行榜展示。

### 为什么选择 AWD Arena？

| 特性 | AWD Arena | 传统平台 |
|------|-----------|----------|
| 部署复杂度 | 单二进制文件 | 多容器/依赖 |
| 性能 | Go高性能后端 | Python较慢 |
| 响应式设计 | 完整支持 | 部分支持 |
| 实时更新 | WebSocket | 轮询 |
| 安全特性 | WAF + XSS防护 | 基础防护 |
| 移动端支持 | 完整适配 | 不支持 |

---

## 功能特性

### 核心功能

- **比赛管理** - 创建/编辑/删除比赛，支持多种模式，自定义轮次，状态机管理
- **队伍系统** - 队伍注册管理，口令加入，队伍-比赛关联
- **计分系统** - 实时攻击分/防御分计算，零和计分算法，首杀奖励
- **Docker容器** - 自动为每个队伍生成独立靶机，容器监控，一键重启
- **监控审计** - 实时攻击态势大屏，完整审计日志，系统健康监控

### 安全特性

- **权限控制** - 基于角色的访问控制（RBAC）
- **XSS防护** - WAF + 输入验证
- **密码策略** - 强制修改默认密码，强度验证
- **路由守卫** - 前端权限验证
- **审计日志** - 所有操作可追溯

### 用户体验

- **响应式设计** - 完美支持桌面和移动端
- **实时更新** - WebSocket 推送
- **现代化UI** - 基于 Ant Design

---

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Docker 20+ (可选)
- SQLite / PostgreSQL / MySQL

### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/your-org/awd-arena.git
cd awd-arena

# 2. 构建后端
go build -o awd-arena ./cmd/server

# 3. 构建前端
cd web && npm install && npm run build && cd ..

# 4. 启动服务
./awd-arena
```

### 访问系统

启动后访问：http://localhost:8080

默认管理员账号：
- 用户名：admin
- 密码：admin123

**首次登录后请立即修改默认密码！**

---

## 使用指南

### 用户角色

| 角色 | 权限 | 说明 |
|------|------|------|
| admin | 全部权限 | 系统管理员 |
| organizer | 比赛管理 | 比赛组织者/裁判 |
| player | 参赛 | 普通参赛选手 |

### 比赛管理

#### 创建比赛

1. 以管理员身份登录
2. 进入 **管理后台 - 比赛管理**
3. 点击 **创建比赛**
4. 填写比赛信息并保存

#### 比赛生命周期

准备 - 进行中 - 暂停 - 结束

### 队伍管理

1. 管理员创建队伍并设置口令
2. 选手使用口令加入队伍
3. 首次登录需修改密码

### 排行榜系统

- 实时更新排名
- 攻击分/防御分/总分
- 首杀标识
- 支持导出 PDF/CSV

---

## API文档

### 认证 API

登录: POST /api/v1/auth/login
注册: POST /api/v1/auth/register
修改密码: PUT /api/v1/auth/change-password
登出: POST /api/v1/auth/logout

### 比赛管理 API

获取比赛列表: GET /api/v1/games
创建比赛: POST /api/v1/games
获取比赛详情: GET /api/v1/games/:id
更新比赛: PUT /api/v1/games/:id
删除比赛: DELETE /api/v1/games/:id
启动比赛: POST /api/v1/games/:id/start
暂停比赛: POST /api/v1/games/:id/pause

### Flag API

提交Flag: POST /api/v1/flags/submit

### 排行榜 API

获取排行榜: GET /api/v1/games/:id/rankings

### 容器管理 API

获取容器列表: GET /api/v1/games/:id/containers
重启容器: POST /api/v1/games/:id/containers/:cid/restart

---

## 部署指南

### systemd 服务

创建服务文件 /etc/systemd/system/awd-arena.service

启动: systemctl start awd-arena
状态: systemctl status awd-arena
开机自启: systemctl enable awd-arena

### Nginx 反向代理

配置 WebSocket 和静态文件代理

### 环境变量

DB_TYPE, DB_HOST, JWT_SECRET, SERVER_PORT

---

## 常见问题

Q: 忘记管理员密码？ - 直接修改数据库
Q: 容器无法启动？ - 检查 Docker 服务
Q: WebSocket 连接失败？ - 检查代理配置
Q: 排行榜不更新？ - 确保 WebSocket 连接正常

---

## 安全建议

- 生产环境必须修改默认密码
- 使用强 JWT 密钥
- 启用 HTTPS
- 定期备份数据库
- 限制管理后台访问 IP

---

## 许可证

MIT License

---

Made with love by ClawX Team

