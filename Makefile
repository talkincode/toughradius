.PHONY: help build build-backend buildf runs runf dev clean test test-integration test-integration-pg test-eap-acceptance-docker initdb killfs version lint ci setup-hooks

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "develop")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

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
	@echo "  make build      - 构建完整版本 (前端+后端，推荐)"
	@echo "  make buildf     - 仅构建前端"
	@echo "  make build-backend - 仅构建后端（假设前端已构建）"
	@echo ""
	@echo "Quality:"
	@echo "  make test       - 运行测试"
	@echo "  make test-eap-acceptance-docker - 在 Linux 容器中运行外部 EAP 验收测试"
	@echo "  make lint       - 运行代码检查"
	@echo "  make ci         - 运行完整 CI 检查（本地）"
	@echo "  make setup-hooks - 安装 Git hooks"
	@echo ""
	@echo "Database:"
	@echo "  make initdb     - 初始化数据库（危险操作，会删除所有数据）"
	@echo ""
	@echo "Maintenance:"
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
build: buildf
	@echo ""
	@echo "🔨 构建后端生产版本..."
	@echo "📦 Version: $(VERSION)"
	@echo "🕐 Build Time: $(BUILD_TIME)"
	@echo "📝 Git Commit: $(GIT_COMMIT)"
	@echo "⚠️  Static build (CGO_ENABLED=0)"
	@echo ""
	@echo "🔍 验证前端构建..."
	@test -f web/dist/admin/index.html || (echo "❌ 错误: web/dist/admin/index.html 不存在！" && exit 1)
	@test -d web/dist/admin/assets || (echo "❌ 错误: web/dist/admin/assets 目录不存在！" && exit 1)
	@ASSET_COUNT=$$(find web/dist/admin/assets -type f 2>/dev/null | wc -l | tr -d ' '); \
	if [ "$$ASSET_COUNT" -lt 1 ]; then \
		echo "❌ 错误: web/dist/admin/assets 中没有文件！"; \
		exit 1; \
	fi; \
	echo "✅ 前端验证通过 ($$ASSET_COUNT 个资源文件)"
	@echo ""
	@mkdir -p release
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o release/toughradius main.go
	@echo ""
	@SIZE=$$(ls -lh release/toughradius | awk '{print $$5}'); \
	echo "✅ 构建完成: release/toughradius ($$SIZE)"
	@echo "📁 前端已嵌入二进制文件"

# 仅构建后端（不重新构建前端，假设前端已存在）
build-backend:
	@echo "🔨 仅构建后端（跳过前端构建）..."
	@echo "📦 Version: $(VERSION)"
	@echo "🕐 Build Time: $(BUILD_TIME)"
	@echo "📝 Git Commit: $(GIT_COMMIT)"
	@if [ ! -f web/dist/admin/index.html ]; then \
		echo "⚠️  警告: 前端未构建，正在构建..."; \
		$(MAKE) buildf; \
	fi
	@mkdir -p release
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o release/toughradius main.go
	@SIZE=$$(ls -lh release/toughradius | awk '{print $$5}'); \
	echo "✅ 构建完成: release/toughradius ($$SIZE)"

# 显示版本信息
version:
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# 构建前端生产版本
buildf:
	@echo "🎨 构建前端生产版本..."
	@if [ ! -d web/node_modules ]; then \
		echo "📦 安装前端依赖..."; \
		cd web && npm ci; \
	fi
	@echo "🔍 运行 TypeScript 类型检查..."
	@cd web && npm run type-check
	@echo "🏗️  编译前端资源..."
	@cd web && npm run build
	@echo ""
	@ASSET_COUNT=$$(find web/dist/admin/assets -type f 2>/dev/null | wc -l | tr -d ' '); \
	echo "✅ 前端构建完成: web/dist/admin/ ($$ASSET_COUNT 个资源文件)"
	@ls -lh web/dist/admin/index.html web/dist/admin/assets/*.js 2>/dev/null | head -5

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

# 运行基于容器 PostgreSQL 的集成测试（真实数据库）
# 自动拉起 docker-compose.test.yml 中的 Postgres，运行 //go:build integration 套件，最后清理。
test-integration-pg:
	@echo "🐘 启动 PostgreSQL 测试容器并运行集成测试..."
	@set -e; \
	COMPOSE="docker compose -p toughradius-it -f docker-compose.test.yml"; \
	cleanup() { echo "🧹 清理测试容器..."; $$COMPOSE down -v >/dev/null 2>&1 || true; }; \
	trap cleanup EXIT; \
	$$COMPOSE up -d; \
	echo "⏳ 等待 Postgres 健康检查..."; \
	for i in $$(seq 1 30); do \
		status=$$($$COMPOSE ps --format '{{.Health}}' 2>/dev/null || true); \
		[ "$$status" = "healthy" ] && break; \
		sleep 2; \
	done; \
	TEST_DATABASE_HOST=127.0.0.1 \
	TEST_DATABASE_PORT=15432 \
	TEST_DATABASE_USER=toughradius \
	TEST_DATABASE_PASSWORD=toughradius \
	TEST_DATABASE_NAME=postgres \
	INTEGRATION_REQUIRED=1 \
	CGO_ENABLED=0 go test -tags=integration -count=1 -v ./test/integration/...

# 在 Linux 容器中运行外部 eapol_test EAP 验收测试，适合 macOS 本地复现 CI 路径。
# 该目标不会更新 docs/reports；本地 JSON、日志和报告写入 build/eap-acceptance/。
test-eap-acceptance-docker:
	@echo "🔐 在 Linux 容器中运行 EAP 外部验收测试..."
	@set -e; \
	COMPOSE="docker compose -p toughradius-it -f docker-compose.test.yml"; \
	NETWORK="toughradius-it_default"; \
	cleanup() { echo "🧹 清理测试容器..."; $$COMPOSE down -v >/dev/null 2>&1 || true; }; \
	trap cleanup EXIT; \
	$$COMPOSE up -d; \
	echo "⏳ 等待 Postgres 健康检查..."; \
	for i in $$(seq 1 30); do \
		status=$$($$COMPOSE ps --format '{{.Health}}' 2>/dev/null || true); \
		[ "$$status" = "healthy" ] && break; \
		sleep 2; \
	done; \
	mkdir -p build/eap-acceptance; \
	test_status=0; \
	docker run --rm \
		--network "$$NETWORK" \
		-v "$$(pwd):/workspace" \
		-w /workspace \
		-e TEST_DATABASE_HOST=postgres \
		-e TEST_DATABASE_PORT=5432 \
		-e TEST_DATABASE_USER=toughradius \
		-e TEST_DATABASE_PASSWORD=toughradius \
		-e TEST_DATABASE_NAME=postgres \
		-e INTEGRATION_REQUIRED=1 \
		-e EAP_ACCEPTANCE_REQUIRED=1 \
		-e EAP_ACCEPTANCE_RESULT_JSON=/workspace/build/eap-acceptance/eap-acceptance.json \
		-e CGO_ENABLED=0 \
		-e GOCACHE=/tmp/go-build \
		golang:1.25-bookworm \
		bash -c 'set -o pipefail; apt-get update && apt-get install -y --no-install-recommends eapoltest ca-certificates && mkdir -p build/eap-acceptance && go test -tags="integration eap_accept" -count=1 -run TestEAPExternalAcceptance -v ./test/integration/... 2>&1 | tee build/eap-acceptance/go-test.log' || test_status=$$?; \
	if [ -f build/eap-acceptance/eap-acceptance.json ]; then \
		go run ./scripts/eap_acceptance_report.go \
			-input build/eap-acceptance/eap-acceptance.json \
			-report-dir build/eap-acceptance/reports \
			-docs-site-src build/eap-acceptance/docs-site \
			-date "$$(date -u +%F)" \
			-retention 3; \
	fi; \
	echo "✅ EAP 验收输出: build/eap-acceptance/"; \
	exit $$test_status

# 代码检查
lint:
	@echo "🔍 运行代码检查..."
	@echo ""
	@echo "📝 Checking code formatting..."
	@UNFORMATTED=$$(gofmt -l . 2>/dev/null | grep -v vendor || true); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "❌ The following files need formatting:"; \
		echo "$$UNFORMATTED"; \
		echo "Run 'go fmt ./...' to fix"; \
		exit 1; \
	fi
	@echo "✅ Code formatting OK"
	@echo ""
	@echo "🔎 Running go vet..."
	@CGO_ENABLED=0 go vet ./...
	@echo "✅ go vet OK"
	@echo ""
	@if command -v golangci-lint > /dev/null 2>&1; then \
		echo "🔍 Running golangci-lint..."; \
		golangci-lint run --timeout=5m || true; \
	else \
		echo "💡 Tip: Install golangci-lint for more thorough checks:"; \
		echo "   brew install golangci-lint"; \
	fi
	@echo ""
	@echo "✅ Lint checks completed"

# 本地 CI 检查（模拟 GitHub Actions）
ci: lint test build
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ All CI checks passed!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 安装 Git hooks
setup-hooks:
	@echo "🔧 Setting up Git hooks..."
	@chmod +x .githooks/pre-commit .githooks/pre-push
	@git config core.hooksPath .githooks
	@echo "✅ Git hooks installed!"
	@echo ""
	@echo "📋 Installed hooks:"
	@echo "   • pre-commit: 格式检查、go vet、快速构建"
	@echo "   • pre-push:   完整测试、lint、构建验证"
	@echo ""
	@echo "💡 To disable hooks temporarily:"
	@echo "   git commit --no-verify"
	@echo "   git push --no-verify"

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
