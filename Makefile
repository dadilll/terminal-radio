
APP_NAME = radio
MAIN_FILE = ./cmd/main.go

.PHONY: all build run check clean

all: build

check:
	@command -v mpv >/dev/null 2>&1 || { \
		echo "❌ mpv is not installed or not in PATH. Please install it."; exit 1; \
	}
	@command -v go >/dev/null 2>&1 || { \
		echo "❌ Go is not installed or not in PATH. Please install it."; exit 1; \
	}
	@echo "✅ All dependencies are installed."

build: check
	@echo "🔧 Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build complete."

run: build
	@echo "🚀 Running $(APP_NAME)..."
	./$(APP_NAME)

clean:
	@echo "🧹 Cleaning up..."
	rm -f $(APP_NAME)
