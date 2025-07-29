# Makefile for BinanceFeederV2 Plugin

# 变量定义
PLUGIN_NAME = binancefeeder_v2.so
PLUGIN_DIR = contrib/binancefeeder_v2
BUILD_DIR = $(PLUGIN_DIR)/build
GO = go
GOFLAGS = -buildmode=plugin

# 默认目标
.PHONY: all
all: build

# 创建构建目录
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# 构建插件
.PHONY: build
build: $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) .

# 安装插件到GOPATH
.PHONY: install
install: build
	cp $(BUILD_DIR)/$(PLUGIN_NAME) $(GOPATH)/bin/

# 清理构建文件
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

# 测试
.PHONY: test
test:
	$(GO) test ./$(PLUGIN_DIR)/...

# 格式化代码
.PHONY: fmt
fmt:
	$(GO) fmt ./$(PLUGIN_DIR)/...

# 代码检查
.PHONY: vet
vet:
	$(GO) vet ./$(PLUGIN_DIR)/...

# 运行示例
.PHONY: example
example: build
	@echo "Starting MarketStore with BinanceFeederV2 plugin..."
	@echo "Make sure to update mkts.example.yml with your API keys"
	marketstore start --config $(PLUGIN_DIR)/mkts.example.yml

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build    - Build the plugin"
	@echo "  install  - Install the plugin to GOPATH/bin"
	@echo "  clean    - Clean build files"
	@echo "  test     - Run tests"
	@echo "  fmt      - Format code"
	@echo "  vet      - Run go vet"
	@echo "  example  - Run example with MarketStore"
	@echo "  help     - Show this help message"

# 开发模式：自动重新构建
.PHONY: dev
dev:
	@echo "Watching for changes and rebuilding..."
	@while true; do \
		$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(PLUGIN_NAME) .; \
		sleep 2; \
	done

# 检查依赖
.PHONY: deps
deps:
	$(GO) mod tidy
	$(GO) mod download

# 生成文档
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@echo "# BinanceFeederV2 Plugin" > $(PLUGIN_DIR)/README.md
	@echo "" >> $(PLUGIN_DIR)/README.md
	@echo "## Features" >> $(PLUGIN_DIR)/README.md
	@echo "- Real-time data streaming via WebSocket" >> $(PLUGIN_DIR)/README.md
	@echo "- Historical data backfill" >> $(PLUGIN_DIR)/README.md
	@echo "- Support for multiple timeframes" >> $(PLUGIN_DIR)/README.md
	@echo "- Data aggregation" >> $(PLUGIN_DIR)/README.md
	@echo "" >> $(PLUGIN_DIR)/README.md
	@echo "## Configuration" >> $(PLUGIN_DIR)/README.md
	@echo "See mkts.example.yml for configuration options." >> $(PLUGIN_DIR)/README.md 