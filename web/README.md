# ToughRADIUS v9 Web UI

基于 React Admin 构建的 ToughRADIUS v9 管理界面。

## 技术栈

- **React 18** - UI 框架
- **TypeScript 5** - 类型安全
- **React Admin 5** - 管理界面框架
- **Vite 5** - 构建工具
- **ECharts** - 数据可视化

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview

# 类型检查
npm run type-check

# 代码检查
npm run lint
```

## 项目结构

```
src/
├── main.tsx              # 应用入口
├── App.tsx               # 主应用组件
├── providers/            # React Admin 提供器
│   ├── dataProvider.ts   # 数据提供器
│   └── authProvider.ts   # 认证提供器
├── pages/                # 页面组件
│   └── Dashboard.tsx     # 仪表盘
├── resources/            # 资源组件（列表、编辑、详情）
└── components/           # 通用组件
```

## 静态编译

构建后的静态文件将嵌入到 Go 二进制文件中，实现单文件部署。

```bash
# 构建前端
npm run build

# Go 将自动嵌入 dist/ 目录
cd ..
go build
```

## API 代理配置

开发模式下，Vite 会将 `/api` 请求代理到后端服务器（默认 `http://localhost:1816`）。

生产模式下，前端和后端服务在同一个端口。
