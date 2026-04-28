BIN			 := codex-web
OUTPUT_DIR	 := bin
PLATFORMS	 := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

CURRENT_TIME := $(shell date "+%Y%m%d_%H%M%S")

clean:
	@echo "🧹清理构建文件..."
	@if [ -d $(OUTPUT_DIR) ]; then rm -rf $(OUTPUT_DIR); fi

build: clean
	@echo "🔖测试版本: $(CURRENT_TIME)"
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(printf '%s' $$platform | cut -d/ -f1); \
		GOARCH=$$(printf '%s' $$platform | cut -d/ -f2); \
		GOARM=$$(printf '%s' $$platform | cut -d/ -f3); \
		if [ "$$GOARM" = "$$platform" ]; then GOARM=""; fi; \
		GOARM=$${GOARM#v}; \
		EXT=""; \
		if [ "$$GOOS" = "windows" ]; then EXT=".exe"; fi; \
		OUTPUT_FILE="$(OUTPUT_DIR)/$(BIN)-$$GOOS-$$GOARCH$${GOARM:+v$$GOARM}$$EXT"; \
		echo "🔨编译 $$OUTPUT_FILE..."; \
		env CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH GOARM=$$GOARM go build -trimpath -ldflags "-X 'main.Version=$(CURRENT_TIME)' -s -w" -o $$OUTPUT_FILE main.go || exit 1; \
		du -sh $$OUTPUT_FILE || exit 1; \
	done

.DEFAULT_GOAL := build
