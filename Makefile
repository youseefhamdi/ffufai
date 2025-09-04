# Makefile for ffufai - AI-powered ffuf wrapper

# Variables
BINARY_NAME = ffufai
VERSION = 1.0.0
BUILD_DIR = build
SOURCE_FILE = ffufai-improved.go
INSTALL_DIR = $(HOME)/.local/bin

# Go build flags
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -w -s"
BUILD_FLAGS = -trimpath

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "🔨 Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_FILE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building for Linux AMD64..."
	@GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SOURCE_FILE)
	
	@echo "Building for Linux ARM64..."
	@GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(SOURCE_FILE)
	
	@echo "Building for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SOURCE_FILE)
	
	@echo "Building for macOS ARM64 (Apple Silicon)..."
	@GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(SOURCE_FILE)
	
	@echo "Building for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SOURCE_FILE)
	
	@echo "✅ Multi-platform build complete"
	@ls -la $(BUILD_DIR)/

# Install the binary to local bin directory
.PHONY: install
install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Installed to $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo ""
	@echo "💡 Make sure $(INSTALL_DIR) is in your PATH"
	@echo "   Add to ~/.bashrc or ~/.zshrc: export PATH=\"\$$PATH:$(INSTALL_DIR)\""

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Uninstalled $(BINARY_NAME)"

# Run tests (placeholder for future tests)
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "✅ Clean complete"

# Format Go code
.PHONY: fmt
fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint the code
.PHONY: lint
lint:
	@echo "🔍 Linting code..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Please install golangci-lint"; exit 1; }
	@golangci-lint run
	@echo "✅ Linting complete"

# Run the binary directly
.PHONY: run
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Show help
.PHONY: help
help:
	@echo "🔧 ffufai Makefile"
	@echo "=================="
	@echo ""
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  install    - Install to ~/.local/bin"
	@echo "  uninstall  - Remove from ~/.local/bin"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  fmt        - Format Go code"
	@echo "  lint       - Lint the code"
	@echo "  run        - Build and run with ARGS"
	@echo "  help       - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make run ARGS='--version'"
	@echo "  make run ARGS='--dry-run -u https://example.com/FUZZ -w /dev/null'"

# Check prerequisites
.PHONY: check
check:
	@echo "🔍 Checking prerequisites..."
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is not installed"; exit 1; }
	@echo "✅ Go: $$(go version)"
	@command -v ffuf >/dev/null 2>&1 || { echo "⚠️  ffuf is not installed (optional)"; }
	@command -v ffuf >/dev/null 2>&1 && echo "✅ ffuf: $$(ffuf -V 2>/dev/null | head -1)" || true
	@[ -n "$$PERPLEXITY_API_KEY" ] && echo "✅ PERPLEXITY_API_KEY is set" || echo "⚠️  PERPLEXITY_API_KEY is not set"

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "🛠️  Setting up development environment..."
	@go mod tidy
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	}
	@command -v ffuf >/dev/null 2>&1 || { \
		echo "Installing ffuf..."; \
		go install github.com/ffuf/ffuf@latest; \
	}
	@echo "✅ Development environment ready"

# Create release archives
.PHONY: release
release: build-all
	@echo "📦 Creating release archives..."
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [[ $$binary == *.exe ]]; then \
			zip release/$${binary}.zip $$binary ../README.md ../LICENSE 2>/dev/null || zip release/$${binary}.zip $$binary ../README.md; \
		else \
			tar -czf release/$${binary}.tar.gz $$binary ../README.md ../LICENSE 2>/dev/null || tar -czf release/$${binary}.tar.gz $$binary ../README.md; \
		fi; \
	done
	@echo "✅ Release archives created in $(BUILD_DIR)/release/"
	@ls -la $(BUILD_DIR)/release/

# Show version
.PHONY: version
version:
	@echo "ffufai version $(VERSION)"

.DEFAULT_GOAL := help