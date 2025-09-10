# Claude2Api+

Claude2Api+ 是一个 OpenAI 兼容的 Claude 代理服务，内置智能 Session 池、健康度调度、熔断与重试，并提供可视化管理面板。

## 核心特性

- 智能 Session 管理：轮询/健康度优先/加权/自适应调度，自动故障转移
- 高可用与稳定性：熔断器、冷却期、错误分类与指数退避重试
- OpenAI 兼容：`/v1/chat/completions`、`/v1/models` 接口格式兼容
- 管理面板：会话管理、运行统计、配置更新（支持 WebSocket 实时推送）
- 安全加固：管理端 JWT 保护、敏感日志脱敏、CORS/WS 来源白名单

## 快速开始

1) 准备配置

- 复制示例配置：`cp config.example.yaml config.yaml`
- 编辑 `config.yaml`：填入你的 Claude `sessions`、`apiKey`、可选 `corsAllowedOrigins`

2) 运行后端

- 开发模式：`go run main.go`
- 或构建：`go build -o server . && ./server`

3) 运行前端（可选）

- `cd frontend && npm install && npm run dev`
- 打开 `http://localhost:3000`

## 鉴权与路由

- 公开：`GET /health`
- 业务 API（需 API Key）：
  - `POST /v1/chat/completions`
  - `GET /v1/models`
- WebSocket（需 API Key）：
  - `GET /ws?token=<APIKEY>`
- 管理端（JWT）：
  - 公开 `POST /admin/login` 获取 token
  - 需 JWT：`GET /admin/me`、`/admin/sessions*`、`/admin/stats`、`/admin/config`

说明：为便于迁移，管理端暂时兼容使用与服务端相同的 API Key 访问（当 JWT 无效时）。建议前端尽快统一切换到 JWT，随后可关闭该兼容。

## 配置项（config.yaml）

- `sessions`：Session 列表（`sessionKey`、可选 `orgID`）
- `sessionManager`：调度策略、健康检查、熔断、最大重试、冷却期
- `address`：监听地址（默认 `0.0.0.0:8080`）
- `apiKey`：业务 API 的访问密钥
- `proxy`：上游代理
- `chatDelete`：是否自动删除会话
- `maxChatHistoryLength`：大上下文阈值
- `enableMirrorApi` / `mirrorApiPrefix`：镜像接口（可选）
- `adminUser` / `adminPassword` / `adminSecret`：管理端用户名/密码/JWT 密钥
- `corsAllowedOrigins`：允许跨域来源（数组），默认 `*`，生产建议显式列出域名

环境变量等价项：`SESSIONS`、`APIKEY`、`CORS_ORIGINS`、`SESSION_MANAGER_*` 等，详见 `config/config.go`。

## 前后端对接说明

- 业务 API：前端请求需设置 `Authorization: Bearer <APIKEY>`
- 管理端：先 `POST /admin/login` 获取 JWT，再以 `Authorization: Bearer <JWT>` 访问 `/admin/*`
- WebSocket：浏览器原生无法加自定义头，故以查询参数携带：`/ws?token=<APIKEY>`

## 安全建议

- 生产环境修改默认管理员凭据与 `adminSecret`
- 将 `corsAllowedOrigins` 设置为受信域名，避免默认 `*`
- 逐步移除管理端 API Key 兼容逻辑，统一使用 JWT

## 开发与构建

- Go：`go 1.22+`（本项目在 go1.25 测试通过）
- 前端：Node 18+，`npm install && npm run dev`

## 许可

本项目遵循仓库内 `LICENSE`。