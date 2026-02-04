# EnterAI 部署指南

本文档介绍如何部署改造后的 Chatbox EnterAI 版本。

## 架构概述

```
┌─────────────────────────────────────────────────┐
│                    用户浏览器                    │
│                  http://host:9980               │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│              Nginx (chatbox-web)                 │
│              外部端口: 9980                      │
│              ┌─────────────────────────────┐    │
│              │  /api/*  → 反向代理到后端    │    │
│              │  /*      → 静态文件/SPA     │    │
│              └─────────────────────────────┘    │
└─────────────────────────────────────────────────┘
                        │
                        │ Docker 内部网络
                        ▼
┌─────────────────────────────────────────────────┐
│              Go 后端 (chatbox-backend)           │
│              内部端口: 8080 (不对外暴露)         │
│              - 用户认证 API                      │
│              - Provider 配置 API                 │
│              - SQLite 数据库                     │
└─────────────────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────┐
│              ./data/chatbox.db                   │
│              SQLite 数据持久化                   │
└─────────────────────────────────────────────────┘
```

**特点**：
- 后端不对外暴露端口，只在 Docker 内部网络通信
- 所有 API 请求通过 Nginx 反向代理转发
- 只需开放一个端口 (9980)

## 快速开始

### 1. 使用 Docker Compose 部署

```bash
# 创建 .env 文件（可选，配置 JWT 密钥）
cp .env.example .env

# 启动服务
docker compose up -d

# 查看日志
docker compose logs -f
```

服务启动后：
- Web 界面: http://localhost:9980
- API 服务: http://localhost:9980/api (通过 Nginx 反向代理)

### 2. 首次使用

1. 访问 http://localhost:9980
2. 点击右上角的登录按钮，或访问设置页面
3. **第一个注册的用户会自动成为管理员**
4. 管理员可以在设置中配置系统级 Provider

### 3. 管理员配置 Provider

1. 使用管理员账户登录
2. 进入 设置 → EnterAI → 管理 Providers
3. 添加 Provider 配置：
   - Provider ID: `enter-ai`
   - 名称: `EnterAI`
   - API 风格: `openai` / `google` / `anthropic`
   - API Host: 你的 AI 服务地址
   - API Key: 系统级 API 密钥
   - 模型列表: JSON 格式的模型配置

示例模型配置：
```json
[
  {
    "modelId": "gpt-4o",
    "nickname": "GPT-4o",
    "capabilities": ["vision", "tool_use"]
  },
  {
    "modelId": "gpt-4o-mini",
    "nickname": "GPT-4o Mini"
  }
]
```

## 数据库说明

- 数据库文件位于 `./data/chatbox.db`
- 首次启动时会自动执行 `backend/migrations/` 目录下的 SQL 迁移
- 迁移是幂等的，可以安全地重启服务

### 迁移机制

- SQL 文件按数字顺序执行（如 `001_init_schema.sql`, `002_xxx.sql`）
- `schema_migrations` 表记录已执行的迁移版本和 checksum
- 如果已执行的迁移文件被修改，服务会拒绝启动

## 环境变量

| 变量 | 默认值 | 说明 |
|-----|-------|------|
| `JWT_SECRET` | `change-me-in-production` | JWT 签名密钥 |
| `DB_PATH` | `/app/data/chatbox.db` | SQLite 数据库路径 |
| `SERVER_PORT` | `8080` | 后端服务端口 |
| `API_BASE_URL` | `` (空) | 前端 API 地址，生产环境为空（使用 Nginx 代理） |

## API 接口

### 公开接口

| 接口 | 方法 | 说明 |
|-----|------|------|
| `/api/auth/register` | POST | 用户注册 |
| `/api/auth/login` | POST | 用户登录 |
| `/api/config/providers` | GET | 获取可用 Provider |

### 需要认证

| 接口 | 方法 | 说明 |
|-----|------|------|
| `/api/auth/me` | GET | 获取当前用户信息 |

### 管理员接口

| 接口 | 方法 | 说明 |
|-----|------|------|
| `/api/admin/providers` | GET/POST | 获取/创建 Provider |
| `/api/admin/providers/:id` | PUT/DELETE | 更新/删除 Provider |
| `/api/admin/users` | GET | 获取用户列表 |

## 开发模式

```bash
# 启动后端（需要 Go 环境）
cd backend
go run .

# 启动前端（需要 Node.js 环境）
npm install
npm run dev:web
```

## 注意事项

1. **生产环境必须修改 JWT_SECRET**
2. 数据库文件应该定期备份
3. 第一个注册用户会成为管理员，请确保管理员账户安全
4. 普通用户无需登录也可以使用系统配置的 Provider
