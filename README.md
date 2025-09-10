# Claude2Api+

一个智能的 Claude API 代理服务，提供 Session 池管理和负载均衡功能。

## 功能特性

- 🔄 **智能 Session 管理**: 自动负载均衡和故障转移
- 📊 **实时监控**: Session 健康状态和系统统计
- 🎯 **高可用性**: 熔断器和重试机制
- 🌐 **OpenAI 兼容**: 完全兼容 OpenAI API 格式
- 💻 **管理界面**: 现代化的 Web 管理面板

## 快速开始

### 后端启动

```bash
# 方法1：使用配置文件
cp config.example.yaml config.yaml
# 编辑 config.yaml 配置你的 Session 和 API Key
go run main.go

# 方法2：使用环境变量
export SESSIONS="sk-ant-sid01-your-session-key:your-org-id"
export APIKEY="your-api-key-here"
go run main.go
```

### 前端管理界面

```bash
cd frontend
npm install
npm run dev
```

访问 `http://localhost:3000` 打开管理界面。

## 配置说明

主要配置项：

- `sessions`: Claude Session 密钥列表
- `port`: 服务端口
- `max_retries`: 最大重试次数
- `timeout`: 请求超时时间
- `circuit_breaker`: 熔断器配置

## API 端点

- `GET /health` - 健康检查
- `GET /admin/sessions` - Session 状态
- `GET /admin/stats` - 系统统计
- `POST /v1/chat/completions` - 聊天接口

## 管理界面功能

- **Session 管理**: 添加/删除 Session 密钥
- **状态监控**: 实时健康状态和性能指标
- **系统统计**: 请求量和响应时间统计
- **配置管理**: API 密钥和系统设置

## 技术栈

- **后端**: Go + Gin
- **前端**: React + TypeScript + Vite
- **样式**: Tailwind CSS
- **图表**: Chart.js
- **状态管理**: Zustand