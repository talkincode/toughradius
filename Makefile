PHONY: help build runs runf dev clean test initdb killfs

# 默认目标
help:
	@echo "ToughRADIUS v9 Makefile Commands"
	@echo "================================="
	@echo "Development:"
	@echo "  make runs       - 启动后端服务 (支持 SQLite)"
	@echo "  make runf       - 启动前端开发服务"
	@echo "  make dev        - 同时启动前后端服务"
	@echo "  make killfs     - 停止前后端所有服务"
	@echo ""
	@echo "Build:"
	@echo "  make build      - 构建生产版本 (PostgreSQL only)"
	@echo "  make buildf     - 构建前端生产版本"
	@echo ""
	@echo "Database:"
	@echo "  make initdb     - 初始化数据库（危险操作，会删除所有数据）"
	@echo ""
	@echo "Maintenance:"
	@echo "  make test       - 运行测试"
	@echo "  make clean      - 清理构建文件"
	@echo ""

# 启动后端服务（开发模式，支持 SQLite）
runs:
	@echo "🚀 启动 ToughRADIUS 后端服务..."
	@echo "📝 配置文件: toughradius.yml"
	@echo "🔧 SQLite 支持: 已启用 (CGO_ENABLED=0)"
	@echo ""
	CGO_ENABLED=0 go run main.go -c toughradius.yml

# 启动前端开发服务
runf:
	@echo "🎨 启动前端开发服务..."
	@echo "📂 工作目录: web/"
	@echo "🌐 访问地址: http://localhost:3000/admin"
	@echo ""
	cd web && npm run dev

# 同时启动前后端（需要 tmux 或在不同终端运行）
dev:
	@echo "⚠️  请在两个不同的终端窗口运行："
	@echo "   终端1: make runs"
	@echo "   终端2: make runf"
	@echo ""
	@echo "或使用以下命令在后台运行："
	@echo "   make runs > /tmp/toughradius-backend.log 2>&1 &"
	@echo "   make runf > /tmp/toughradius-frontend.log 2>&1 &"

# 构建生产版本（静态编译，支持 PostgreSQL 和 SQLite）
build:
	@echo "🔨 构建生产版本..."
	@echo "⚠️  Static build (CGO_ENABLED=0)"
	@bash scripts/build-backend.sh

# 构建前端生产版本
buildf:
	@echo "🔨 构建前端生产版本..."
	@cd web && npm run build
	@echo "✅ 前端构建完成: web/dist/"

# 初始化数据库（危险操作）
initdb:
	@echo "⚠️  警告：此操作将删除并重建所有数据库表！"
	@read -p "确认继续？(yes/no): " confirm && [ "$$confirm" = "yes" ] || (echo "已取消"; exit 1)
	@echo "🗄️  初始化数据库..."
	CGO_ENABLED=0 go run main.go -initdb -c toughradius.yml

# 运行测试
test:
	@echo "🧪 运行测试..."
	CGO_ENABLED=0 go test ./...

# 运行集成测试
test-integration:
	@echo "🧪 运行集成测试..."
	CGO_ENABLED=0 go test -v ./internal/radiusd/... -run TestRadiusIntegration

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -rf release/
	rm -rf web/dist/
	rm -f /tmp/toughradius-test
	@echo "✅ 清理完成"

# 安装前端依赖
install-frontend:
	@echo "📦 安装前端依赖..."
	cd web && npm install

# 检查代码格式
fmt:
	@echo "📝 格式化 Go 代码..."
	go fmt ./...
	@echo "📝 格式化前端代码..."
	cd web && npm run format || echo "提示: 如需格式化前端代码，请在 package.json 中添加 format 脚本"

# 查看后端日志
logs:
	@tail -f /tmp/toughradius.log

# 查看前端日志
logsf:
	@tail -f /tmp/frontend.log

# 停止前后端所有服务
killfs:
	@echo "🛑 停止前后端所有服务..."
	@pkill -f "go run main.go" 2>/dev/null || true
	@pkill -f "toughradius" 2>/dev/null || true
	@pkill -f "vite" 2>/dev/null || true
	@pkill -f "npm run dev" 2>/dev/null || true
	@echo "✅ 所有服务已停止"

# 重启后端服务
restart-backend: killfs
	@echo "🔄 重启后端服务..."
	@make runs

# 快速启动（后台运行前后端）
quick-start: killfs
	@echo "🚀 快速启动前后端服务（后台运行）..."
	@make runs > /tmp/toughradius-backend.log 2>&1 &
	@sleep 3
	@make runf > /tmp/toughradius-frontend.log 2>&1 &
	@sleep 2
	@echo ""
	@echo "✅ 服务已启动！"
	@echo "📊 后端: http://localhost:1816"
	@echo "🎨 前端: http://localhost:3000/admin"
	@echo "📝 后端日志: tail -f /tmp/toughradius-backend.log"
	@echo "📝 前端日志: tail -f /tmp/toughradius-frontend.log"
	@echo ""
	@echo "🛑 停止服务: make killfs"
