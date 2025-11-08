# ToughRADIUS v9 - React Admin 界面重构

## 重构概述

使用 React Admin 框架重构 ToughRADIUS 的 Web 管理界面，采用静态编译模式将前端资源嵌入到 Go 二进制文件中。

## 技术栈

### 前端

- **React 18** - UI 框架
- **TypeScript 5** - 类型安全
- **React Admin 5** - 管理界面框架
- **Vite 5** - 构建工具
- **ECharts** - 数据可视化

### 后端

- **Go 1.21+** - 后端语言
- **Echo v4** - Web 框架
- **embed** - 静态资源嵌入

## 项目结构

```
toughradius/
├── web/                          # React Admin 前端
│   ├── src/
│   │   ├── main.tsx             # 应用入口
│   │   ├── App.tsx              # 主应用组件
│   │   ├── providers/           # React Admin 提供器
│   │   │   ├── dataProvider.ts  # 数据提供器
│   │   │   └── authProvider.ts  # 认证提供器
│   │   ├── pages/               # 页面组件
│   │   │   └── Dashboard.tsx    # 仪表盘
│   │   └── resources/           # 资源组件
│   │       ├── radiusUsers.tsx      # RADIUS 用户
│   │       ├── onlineSessions.tsx   # 在线会话
│   │       ├── accounting.tsx       # 计费记录
│   │       └── radiusProfiles.tsx   # RADIUS 配置
│   ├── dist/                    # 构建输出（嵌入到 Go）
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── build.sh                 # 构建脚本
├── web/static.go                # Go 静态文件嵌入
└── internal/webserver/
    └── server.go                # Web 服务器（已更新）
```

## 开发流程

### 1. 前端开发

```bash
cd web

# 安装依赖
npm install

# 启动开发服务器（带 API 代理）
npm run dev

# 访问 http://localhost:3000
```

开发模式下，Vite 会将 `/api` 请求代理到后端服务器（默认 `http://localhost:1816`）。

### 2. 构建生产版本

```bash
cd web

# 构建前端
npm run build

# 或使用构建脚本
chmod +x build.sh
./build.sh
```

构建产物在 `web/dist/` 目录。

### 3. 编译 Go 二进制

```bash
# 返回项目根目录
cd ..

# 编译（自动嵌入 web/dist/）
go build

# 运行
./toughradius -c toughradius.yml
```

## 功能特性

### 已实现的资源管理

1. **RADIUS 用户管理** (`/admin#/radius/users`)

   - 列表、创建、编辑、查看
   - 字段：用户名、密码、配置文件、状态等

2. **在线会话监控** (`/admin#/radius/online`)

   - 实时查看在线用户
   - 会话详情、流量统计

3. **计费记录查询** (`/admin#/radius/accounting`)

   - 多条件筛选
   - 流量统计、会话时长

4. **RADIUS 配置管理** (`/admin#/radius/profiles`)

   - 配置文件管理
   - 带宽限制、IP 池设置

5. **数据可视化仪表盘** (`/admin#/`)
   - 认证趋势图
   - 在线用户分布
   - 流量统计

### React Admin 特性

- ✅ 自动生成 CRUD 界面
- ✅ 响应式设计
- ✅ 国际化支持（中文）
- ✅ 主题定制
- ✅ 权限控制
- ✅ 搜索和过滤
- ✅ 批量操作
- ✅ 数据导出

## API 接口要求

React Admin 需要后端提供标准的 RESTful API：

### 数据格式

```typescript
// 列表接口
GET /api/v1/radius/users?page=1&perPage=10&sort=id&order=DESC
Response: {
  data: [...],
  total: 100
}

// 单条数据
GET /api/v1/radius/users/:id
Response: {
  data: {...}
}

// 创建
POST /api/v1/radius/users
Body: {...}
Response: { id: "new-id" }

// 更新
PUT /api/v1/radius/users/:id
Body: {...}
Response: {...}

// 删除
DELETE /api/v1/radius/users/:id
Response: {...}
```

### 认证接口

```typescript
// 登录
POST /api/v1/auth/login
Body: { username: "admin", password: "password" }
Response: {
  token: "jwt-token",
  permissions: [...]
}
```

## 路由说明

### 前端路由（SPA）

所有 `/admin` 路径下的请求由 React Admin 处理：

- `/admin` - 仪表盘
- `/admin#/radius/users` - RADIUS 用户
- `/admin#/radius/online` - 在线会话
- `/admin#/radius/accounting` - 计费记录
- `/admin#/radius/profiles` - RADIUS 配置

### 后端路由

- `/api/*` - API 接口
- `/admin/*` - React Admin SPA（从嵌入的静态文件服务）
- `/admin/assets/*` - 静态资源（JS、CSS、图片等）
- `/` - 重定向到 `/admin`

### 旧版界面（向后兼容）

- `/static/*` - 旧版静态资源
- `/login` - 旧版登录页面
- 其他模板路由...

## 部署说明

### 单二进制部署

```bash
# 1. 构建前端
cd web && npm run build && cd ..

# 2. 编译 Go（嵌入前端）
go build -o toughradius

# 3. 部署单个二进制文件
scp toughradius user@server:/usr/local/bin/
```

### Docker 部署

```dockerfile
FROM node:18 AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.21 AS go-builder
WORKDIR /app
COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
RUN go build -o toughradius

FROM debian:bullseye-slim
COPY --from=go-builder /app/toughradius /usr/local/bin/
CMD ["toughradius"]
```

## 开发指南

### 添加新资源

1. 创建资源组件文件 `web/src/resources/newResource.tsx`
2. 定义 List、Edit、Create、Show 组件
3. 在 `App.tsx` 中注册资源

```tsx
<Resource
  name="resource-name"
  list={ResourceList}
  edit={ResourceEdit}
  create={ResourceCreate}
  show={ResourceShow}
/>
```

### 自定义字段

```tsx
import { TextField, DateField, FunctionField } from "react-admin";

<FunctionField
  label="自定义"
  render={(record) => `${record.field1}-${record.field2}`}
/>;
```

### 主题定制

修改 `web/src/App.tsx` 添加主题配置：

```tsx
import { defaultTheme } from 'react-admin';

const theme = {
  ...defaultTheme,
  palette: {
    primary: {
      main: '#1976d2',
    },
  },
};

<Admin theme={theme} ...>
```

## 注意事项

1. **开发模式 vs 生产模式**

   - 开发：前端独立运行，API 通过代理访问后端
   - 生产：前端嵌入 Go 二进制，通过同一端口访问

2. **API 兼容性**

   - 确保后端 API 遵循 REST 规范
   - 响应格式符合 React Admin 要求

3. **认证流程**

   - JWT token 存储在 localStorage
   - 每次请求自动添加 Authorization header

4. **类型安全**

   - 使用 TypeScript 确保类型安全
   - 定义清晰的接口类型

5. **性能优化**
   - 代码分割（已在 vite.config.ts 配置）
   - 懒加载路由
   - 图片优化

## TODO

- [ ] 完善 API 接口实现
- [ ] 添加更多资源管理页面
- [ ] 实现权限控制
- [ ] 添加国际化支持
- [ ] 性能监控集成
- [ ] E2E 测试
- [ ] 用户文档

## 参考资源

- [React Admin 文档](https://marmelab.com/react-admin/)
- [Vite 文档](https://vitejs.dev/)
- [ECharts 文档](https://echarts.apache.org/)
- [Go embed 文档](https://pkg.go.dev/embed)
