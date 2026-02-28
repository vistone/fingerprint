.PHONY: help test benchmark lint format clean install-tools build

# 默认目标
help:
	@echo "Fingerprint Project - Available Targets"
	@echo "======================================="
	@echo ""
	@echo "Development:"
	@echo "  make test              - Run all tests"
	@echo "  make benchmark         - Run benchmarks"
	@echo "  make lint              - Run code linting (golangci-lint, go vet, gofmt)"
	@echo "  make format            - Format code with gofmt"
	@echo ""
	@echo "Build:"
	@echo "  make build             - Build all example binaries"
	@echo "  make clean             - Remove build artifacts"
	@echo ""
	@echo "Setup:"
	@echo "  make install-tools     - Install development tools"
	@echo ""
	@echo "Documentation:"
	@echo "  make docs              - Generate API documentation"
	@echo ""

# 运行所有测试
test:
	@echo "Running tests..."
	go test ./... -v -race -timeout=5m

# 运行基准测试
benchmark:
	@echo "Running benchmarks..."
	go test ./test -bench=. -benchmem -run=^$ -timeout=30m
	@echo "Benchmarks completed!"

# 代码检查
lint: lint-vet lint-fmt lint-golangci

lint-vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ go vet passed"

lint-fmt:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "❌ Code formatting issues found:"; \
		gofmt -s -d .; \
		exit 1; \
	else \
		echo "✓ Code formatting is correct"; \
	fi

lint-golangci:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./... --timeout=5m; \
		echo "✓ golangci-lint passed"; \
	else \
		echo "⚠ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 代码格式化
format:
	@echo "Formatting code with gofmt..."
	gofmt -s -w .
	@echo "✓ Code formatted"

# 编译所有示例
build:
	@echo "Building examples..."
	mkdir -p build
	@for example in $$(find examples -maxdepth 1 -type d -name '*' ! -name 'examples'); do \
		echo "Building $$(basename $$example)..."; \
		go build -o build/$$(basename $$example) ./$$example; \
	done
	@echo "✓ Build completed"

# 清理构建文件
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/ *.test *.out coverage.* gosec-results.json
	go clean ./...
	@echo "✓ Clean completed"

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing gosec..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Installing go-mod-upgrade..."
	go install github.com/go-mod-upgrade/go-mod-upgrade@latest
	@echo "✓ All tools installed"

# 生成 API 文档
docs:
	@echo "Generating API documentation..."
	go doc -html ./... > docs/api.html
	@echo "✓ Documentation generated at docs/api.html"

# 代码覆盖
coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out -html=coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# 更新依赖
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy
	@echo "✓ Dependencies updated"

# 安全检查
security:
	@echo "Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "⚠ gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# 所有检查
all: lint test benchmark
	@echo ""
	@echo "════════════════════════════════"
	@echo "✅ All checks passed!"
	@echo "════════════════════════════════"

# 快速检查 (用于开发期间)
quick: format lint test
	@echo ""
	@echo "✅ Quick checks passed!"
