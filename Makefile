BINARY_NAME=awd-arena
CLI_NAME=awd-cli
MIGRATOR_NAME=awd-migrator
BUILD_DIR=build
VERSION?=0.1.0
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build-linux build-windows build-all build-cli-linux build-cli-windows \
        build-migrator-linux build-migrator-windows build-frontend clean

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server

build-cli-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_NAME)-linux-amd64 ./cmd/cli

build-cli-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_NAME)-windows-amd64.exe ./cmd/cli

build-migrator-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(MIGRATOR_NAME)-linux-amd64 ./cmd/migrator

build-migrator-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(MIGRATOR_NAME)-windows-amd64.exe ./cmd/migrator

build-all: build-linux build-windows build-cli-linux build-cli-windows build-migrator-linux build-migrator-windows

build-frontend:
	cd web && npm install && npm run build

clean:
	rm -rf $(BUILD_DIR)
