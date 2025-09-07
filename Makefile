# chilix-msg 项目的 Makefile

# Go 相关参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint
GOCOVER=$(GOCMD) tool cover

# 项目参数
BINARY_NAME=chilix-msg
MAIN_PACKAGE=./cmd/chilix-msg
PACKAGES=$(shell $(GOCMD) list ./... | grep -v /vendor/)
COVERAGE_FILE=coverage.out
MODULE_NAME=$(shell $(GOCMD) list -m)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")

# 构建标志
BUILD_FLAGS=-v

# 终端输出颜色
GREEN=\033[0;32m
NC=\033[0m # 无颜色

.PHONY: all clean test coverage lint fmt vet tidy help tag release publish

# 默认目标
all: lint test

# 运行测试
test:
	@echo "${GREEN}正在运行测试...${NC}"
	$(GOTEST) -v ./...

# 运行测试并生成覆盖率报告
coverage:
	@echo "${GREEN}正在运行测试并生成覆盖率报告...${NC}"
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCOVER) -html=$(COVERAGE_FILE)

# 运行代码检查工具
lint:
	@echo "${GREEN}正在运行代码检查...${NC}"
	$(GOLINT) run

# 格式化代码
fmt:
	@echo "${GREEN}正在格式化代码...${NC}"
	$(GOFMT) ./...

# 运行 go vet 静态分析
vet:
	@echo "${GREEN}正在运行 go vet 静态分析...${NC}"
	$(GOVET) ./...

# 更新依赖
tidy:
	@echo "${GREEN}正在整理依赖...${NC}"
	$(GOMOD) tidy

# 基准测试
bench:
	@echo "${GREEN}正在运行基准测试...${NC}"
	$(GOTEST) -bench=. -benchmem ./...

# 安装依赖
deps:
	@echo "${GREEN}正在安装依赖...${NC}"
	$(GOGET) -v -t ./...
	@if ! command -v $(GOLINT) > /dev/null; then \
		echo "正在安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0 \
	fi

# 生成文档
docs:
	@echo "${GREEN}正在生成文档...${NC}"
	@mkdir -p ./docs/api
	@echo "文档生成占位符"

# 创建版本标签
tag:
	@echo "${GREEN}当前版本: $(VERSION)${NC}"
	@echo "${GREEN}请输入新版本号 (例如 v1.0.0):${NC}"
	@read -p " " new_version; \
	git tag -a $$new_version -m "Release $$new_version"; \
	echo "${GREEN}已创建标签: $$new_version${NC}"

# 发布新版本
release: lint test tidy
	@echo "${GREEN}准备发布新版本...${NC}"
	@echo "${GREEN}当前版本: $(VERSION)${NC}"
	@echo "${GREEN}请确保已经提交所有更改并推送到远程仓库${NC}"
	@echo "${GREEN}请输入新版本号 (例如 v1.0.0):${NC}"
	@read -p " " new_version; \
	git tag -a $$new_version -m "Release $$new_version"; \
	git push origin $$new_version; \
	echo "${GREEN}已创建并推送标签: $$new_version${NC}"

# 发布到 pkg.go.dev
publish:
	@echo "${GREEN}正在发布到 pkg.go.dev...${NC}"
	@echo "${GREEN}模块名称: $(MODULE_NAME)${NC}"
	@echo "${GREEN}当前版本: $(VERSION)${NC}"
	@echo "${GREEN}请确保已经推送最新的版本标签到远程仓库${NC}"
	@echo "${GREEN}正在触发 pkg.go.dev 更新...${NC}"
	GOPROXY=https://proxy.golang.org $(GOCMD) list -m $(MODULE_NAME)@$(VERSION)
	@echo "${GREEN}已触发 pkg.go.dev 更新，请稍后访问 https://pkg.go.dev/$(MODULE_NAME) 查看${NC}"

# 显示帮助信息
help:
	@echo "可用的目标:"
	@echo "  all       - 运行代码检查、测试"
	@echo "  test      - 运行测试"
	@echo "  coverage  - 运行测试并生成覆盖率报告"
	@echo "  lint      - 运行代码检查"
	@echo "  fmt       - 格式化代码"
	@echo "  vet       - 运行 go vet 静态分析"
	@echo "  tidy      - 更新依赖"
	@echo "  bench     - 运行基准测试"
	@echo "  deps      - 安装依赖"
	@echo "  docs      - 生成文档"
	@echo "  tag       - 创建新的版本标签"
	@echo "  release   - 发布新版本（创建并推送标签）"
	@echo "  publish   - 发布到 pkg.go.dev"
	@echo "  help      - 显示此帮助信息"