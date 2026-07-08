# Minecraft Manager

基于 Web 的 Minecraft 服务器管理面板，支持通过 RCON 远程管理服务器、查看在线玩家、执行控制台命令，并通过 WebSocket 实时推送服务器日志。

## 功能特性

- **仪表盘** — 查看服务器在线状态、在线玩家数、TPS、版本信息
- **玩家管理** — 查看在线玩家列表，支持踢出、封禁、设置/取消 OP
- **控制台** — 通过 RCON 执行 Minecraft 命令，实时查看命令输出与服务器日志
- **用户认证** — JWT 登录鉴权，登出时通过 Redis 黑名单失效 Token
- **实时通信** — WebSocket 推送日志、玩家进出等事件

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3、TypeScript、Vite、Vant 4、Pinia、Vue Router、Axios |
| 后端 | Go 1.21、Gin、GORM、JWT、Gorilla WebSocket |
| 数据存储 | MySQL 8.0、Redis 7 |
| 游戏服务端 | Fabric（Java 25） |
| 通信协议 | RCON（Minecraft 远程控制台） |

## 项目结构

```
h-minecraft-manage/
├── backend/                 # Go 后端 API 服务
│   ├── config/              # 环境变量配置
│   ├── handler/             # HTTP 请求处理
│   ├── middleware/          # CORS、JWT 中间件
│   ├── model/               # 数据模型
│   ├── pkg/                 # 工具包（JWT、RCON、Redis、日志）
│   ├── router/              # 路由注册
│   ├── service/             # 业务逻辑
│   ├── websocket/           # WebSocket Hub 与日志监听
│   ├── main.go
│   ├── Dockerfile
│   └── .env                 # 本地环境变量（需自行创建）
├── frontend/                # Vue 3 前端
│   ├── src/
│   │   ├── api/             # API 请求封装
│   │   ├── components/      # 公共组件
│   │   ├── stores/          # Pinia 状态管理
│   │   ├── views/           # 页面（登录、仪表盘、玩家、控制台）
│   │   └── router/          # 前端路由
│   ├── Dockerfile
│   └── nginx.conf.template  # Nginx 反向代理模板
├── minecraft-server/        # Fabric 服务端（Docker 构建上下文）
│   ├── Dockerfile
│   ├── server.properties    # 服务端配置（含 RCON，不提交 Git）
│   ├── mods/                # 模组目录（不提交 Git）
│   ├── world/               # 存档目录（不提交 Git）
│   └── logs/                # 日志目录（不提交 Git）
├── docker-compose.yml       # 一键部署全部服务
├── init.sql                 # 数据库初始化脚本
└── .vscode/                 # VS Code / Cursor 调试配置
```

## 环境要求

- **Docker** 与 **Docker Compose**（推荐，一键部署全套服务）
- 本地开发额外需要：
  - **Go** 1.21+
  - **Node.js** 18+（推荐 LTS）
- Minecraft 服务端需开启 **RCON**

---

## Docker 一键部署（推荐）

### 架构

```
浏览器 → frontend:9001 → backend:8080 → mysql / redis
                              ↓
                         minecraft:25575 (RCON)
                         minecraft/logs   (日志)
```

所有服务由 `docker-compose.yml` 自动加入同一 Docker 网络，通过服务名互相访问（如 `backend`、`minecraft`、`mysql`）。

### 1. 准备 Minecraft 服务端文件

在 `minecraft-server/` 目录下放置已安装好的 Fabric 服务端文件，至少包括：

- `fabric-server-launch.jar`
- `server.jar`
- `libraries/`、`versions/`、`.fabric/`
- `mods/`（如有模组）

可参考 `minecraft-server/start.sh` 在本地生成，再将文件保留在该目录。

### 2. 配置 `server.properties`

确保开启 RCON（密码需与 `backend/.env` 中 `RCON_PASSWORD` 一致）：

```properties
enable-rcon=true
rcon.port=25575
rcon.password=你的RCON密码
server-port=25565
```

同时确认 `eula.txt` 内容为 `eula=true`。

### 3. 配置 `backend/.env`

在 `backend/` 目录创建 `.env` 文件：

```env
SERVER_PORT=8080
JWT_SECRET=minecraft-manager-secret-key-change-in-production
RCON_PASSWORD=你的RCON密码
```

> Docker 部署时，`docker-compose.yml` 会自动覆盖数据库、Redis、RCON 主机名和日志路径，无需在 `.env` 中填写 `DB_HOST`、`RCON_HOST` 等容器内地址。

### 4. 启动全部服务

在项目根目录执行：

```bash
docker compose up -d --build
```

### 5. 访问

| 服务 | 地址 |
|------|------|
| 管理面板 | http://localhost:9001 |
| Minecraft 游戏 | `localhost:25565` |
| MySQL | `localhost:3306` |
| Redis | `localhost:6379` |

默认管理员账号：

| 用户名 | 密码 |
|--------|------|
| `admin` | `admin123` |

### Docker 服务说明

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| `mysql` | mc-mysql | 3306 | 数据库 |
| `redis` | mc-redis | 6379 | Token 黑名单 |
| `minecraft` | mc-minecraft | 25565 | Fabric 服务端 |
| `backend` | mc-backend | 仅容器内 8080 | Go API |
| `frontend` | mc-frontend | 9001→80 | Vue + Nginx |

### 常用 Docker 命令

```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f backend
docker compose logs -f minecraft

# 仅重建并重启后端
docker compose up -d --build backend

# 仅重建并重启 Minecraft 服务端
docker compose up -d --build minecraft

# 停止全部服务
docker compose down
```

> **Windows 注意：** 宿主机 80 端口常被系统占用，前端默认映射到 **9001**。若宿主机 8080 已被占用，compose 中 backend 不映射到宿主机，API 通过前端 Nginx 代理访问。

---

## 本地开发

适合前后端热更新调试，MySQL / Redis 仍用 Docker，Minecraft 可独立运行或使用 Docker。

### 1. 启动基础依赖

```bash
docker compose up -d mysql redis
```

### 2. 配置 `backend/.env`

```env
SERVER_PORT=8080
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=mcuser
DB_PASSWORD=mcpass123
DB_NAME=minecraft_manager
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
JWT_SECRET=minecraft-manager-secret-key-change-in-production
RCON_HOST=127.0.0.1
RCON_PASSWORD=你的RCON密码
LOG_PATH=dev.log
```

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SERVER_PORT` | API 服务端口 | `8080` |
| `DB_HOST` / `DB_PORT` | MySQL 地址与端口 | `127.0.0.1` / `3306` |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | 数据库凭据 | 与 docker-compose 一致 |
| `REDIS_ADDR` / `REDIS_PASSWORD` | Redis 地址与密码 | `127.0.0.1:6379` / 空 |
| `JWT_SECRET` | JWT 签名密钥 | 生产环境务必修改 |
| `RCON_HOST` / `RCON_PASSWORD` | Minecraft RCON 地址与密码 | `127.0.0.1` |
| `LOG_PATH` | 服务器日志路径（实时推送） | `dev.log` |

### 3. 启动后端

```bash
cd backend
go mod download
go run .
```

### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器：`http://localhost:5173`，已配置代理将 `/api` 和 `/ws` 转发到后端。

---

## 前端页面

| 路由 | 页面 | 说明 |
|------|------|------|
| `/login` | 登录 | 用户名密码登录 |
| `/` | 仪表盘 | 服务器状态、TPS、在线人数 |
| `/players` | 玩家管理 | 在线玩家列表与操作 |
| `/console` | 控制台 | 执行命令、查看实时日志 |

## API 接口

所有受保护接口需在请求头携带 `Authorization: Bearer <token>`。

### 认证

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| `POST` | `/api/login` | 用户登录 | 否 |
| `POST` | `/api/logout` | 用户登出 | 是 |
| `GET` | `/api/me` | 获取当前用户信息 | 是 |

### 玩家管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/players` | 获取在线玩家列表 |
| `POST` | `/api/players/kick` | 踢出玩家 `{ "name": "...", "reason": "..." }` |
| `POST` | `/api/players/ban` | 封禁玩家 `{ "name": "...", "reason": "..." }` |
| `POST` | `/api/players/op` | 设置 OP `{ "name": "..." }` |
| `POST` | `/api/players/deop` | 取消 OP `{ "name": "..." }` |

### 控制台与服务器

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/console/exec` | 执行命令 `{ "command": "list" }` |
| `GET` | `/api/server/status` | 获取服务器状态 |

### WebSocket

| 路径 | 说明 |
|------|------|
| `GET /ws` | WebSocket 连接，用于接收实时日志与玩家事件 |

## 本地调试（VS Code / Cursor）

项目已配置 `.vscode/launch.json`，支持两种调试方式：

1. **Debug Backend** — 直接启动后端调试
2. **Debug Backend (启动 Docker 依赖)** — 先启动 MySQL / Redis，再调试后端

按 `F5` 开始调试，需安装 [Go 扩展](https://marketplace.visualstudio.com/items?itemName=golang.go)。环境变量从 `backend/.env` 读取。

## 生产构建

**前端：**

```bash
cd frontend
npm run build
```

构建产物输出到 `frontend/dist/`。Docker 部署时由 Nginx 托管，并反向代理 `/api`、`/ws` 到后端。

**后端：**

```bash
cd backend
CGO_ENABLED=0 go build -o server .
```

## 注意事项

- 生产环境务必修改 `JWT_SECRET`、数据库密码、RCON 密码和默认管理员密码。
- `minecraft-server/` 下的大文件（`libraries/`、`mods/`、`world/`、`*.jar` 等）已加入 `.gitignore`，克隆仓库后需自行准备服务端文件。
- `server.properties` 含敏感配置，不提交 Git，请在本机维护。
- Redis 连接失败时服务仍可启动，但 Token 黑名单功能不可用。
- RCON 连接失败时服务仍可启动，但玩家管理与控制台功能不可用。
- Fabric 新版需要 **Java 25**，`minecraft-server/Dockerfile` 已使用 `eclipse-temurin:25-jdk-jammy`。
- 本地开发时 `LOG_PATH` 建议使用 `backend/dev.log`；Docker 部署时自动挂载 `minecraft-server/logs/latest.log`。

## License

MIT
