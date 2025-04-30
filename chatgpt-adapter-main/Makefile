.PHONY: all build install login-tool clean

all: build

build:
	@echo "Building chatgpt-adapter..."
	go build -o bin/chatgpt-adapter ./main.go
	@echo "Building login tool..."
	go build -o bin/login ./cmd/login/main.go

install: build
	@echo "Installing chatgpt-adapter and login tool..."
	mkdir -p $(GOPATH)/bin
	cp bin/chatgpt-adapter $(GOPATH)/bin/
	cp bin/login $(GOPATH)/bin/
	@echo "Installation complete. You can now use 'chatgpt-adapter' and 'login' commands."

login-tool:
	@echo "Building login tool only..."
	go build -o bin/login ./cmd/login/main.go

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	@echo "Cleanup complete."

help:
	@echo "Available targets:"
	@echo "  all        - Build all components (default)"
	@echo "  build      - Build chatgpt-adapter and login tool"
	@echo "  install    - Build and install to GOPATH/bin"
	@echo "  login-tool - Build only the login tool"
	@echo "  clean      - Remove build artifacts"
	@echo "  help       - Show this help message"

