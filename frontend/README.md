# Claude2Api+ 前端

这是一个现代化的 Claude AI 聊天界面，基于 React 18 + TypeScript + Vite 构建。

## 功能特性

- 🎨 **现代化设计** - 采用 Tailwind CSS，支持深色/浅色主题
- 💬 **实时聊天** - 支持流式响应和思考过程显示
- 🖼️ **图像识别** - 支持拖拽上传图片进行分析
- 📁 **文件管理** - 支持多种文件格式上传
- 🔄 **会话管理** - 完整的对话历史和管理
- 📊 **系统监控** - 实时会话状态和性能监控
- 🎯 **智能重试** - 基于错误类型的智能重试机制
- 📱 **响应式设计** - 完美适配桌面和移动设备

## 技术栈

- **前端框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI 组件**: Tailwind CSS + Headless UI
- **状态管理**: Zustand
- **HTTP 客户端**: Axios
- **路由**: React Router
- **图标**: Lucide React
- **Markdown**: React Markdown
- **文件上传**: React Dropzone

## 快速开始

### 安装依赖

```bash
cd frontend
npm install
```

### 开发环境

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

### 预览生产版本

```bash
npm run preview
```

## 项目结构

```
frontend/
├── public/                 # 静态资源
├── src/
│   ├── components/         # 组件
│   │   ├── ui/            # 基础 UI 组件
│   │   ├── chat/          # 聊天相关组件
│   │   ├── admin/         # 管理后台组件
│   │   └── layout/        # 布局组件
│   ├── pages/             # 页面组件
│   ├── hooks/             # 自定义 Hook
│   ├── stores/            # 状态管理
│   ├── services/          # API 服务
│   ├── utils/             # 工具函数
│   ├── types/             # TypeScript 类型
│   └── styles/            # 样式文件
├── index.html             # HTML 模板
├── package.json           # 依赖配置
├── vite.config.ts         # Vite 配置
├── tailwind.config.js     # Tailwind 配置
└── tsconfig.json          # TypeScript 配置
```

## 主要功能

### 1. 聊天界面
- 实时流式对话
- 支持多种 Claude 模型
- Markdown 渲染
- 代码高亮
- 思考过程显示

### 2. 文件处理
- 拖拽上传图片
- 图片预览和删除
- 多文件上传支持
- 文件大小限制

### 3. 会话管理
- 创建新对话
- 对话历史查看
- 删除对话
- 会话持久化

### 4. 设置管理
- API 密钥配置
- 模型选择
- 连接测试
- 主题切换

### 5. 系统管理
- 会话健康状态监控
- 系统统计信息
- 性能指标展示
- 实时数据更新

## 配置说明

### 环境变量

创建 `.env` 文件：

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_DEFAULT_MODEL=claude-3-7-sonnet-20250219
```

### API 配置

应用启动后，在设置页面配置：
- API 密钥
- 服务器地址
- 默认模型

## 开发指南

### 添加新组件

1. 在 `src/components` 相应目录下创建组件
2. 导出组件并添加 TypeScript 类型
3. 在需要的地方导入使用

### 状态管理

使用 Zustand 进行状态管理：

```typescript
import { create } from 'zustand';

interface StoreState {
  count: number;
  increment: () => void;
}

export const useStore = create<StoreState>((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
}));
```

### API 服务

在 `src/services/api.ts` 中添加新的 API 方法：

```typescript
async customApiMethod(params: CustomParams) {
  const response = await this.api.post('/custom-endpoint', params);
  return response.data;
}
```

### 样式指南

- 使用 Tailwind CSS 类名
- 遵循 BEM 命名约定
- 响应式设计优先
- 支持深色模式

## 部署

### Docker 部署

```dockerfile
FROM node:18-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### 静态文件部署

构建后的文件在 `dist` 目录，可以部署到任何静态文件服务器。

## 浏览器支持

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT License