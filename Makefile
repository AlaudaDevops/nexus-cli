.PHONY: build clean install test fmt vet

BINARY_NAME=nexus-cli
VERSION?=dev
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/alauda/nexus-cli/cmd.Version=$(VERSION) -X github.com/alauda/nexus-cli/cmd.GitCommit=$(GIT_COMMIT) -X github.com/alauda/nexus-cli/cmd.BuildDate=$(BUILD_DATE)"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

clean:
	go clean
	rm -f $(BINARY_NAME)

install:
	go install $(LDFLAGS)

test:
	go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

run:
	go run main.go

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe main.go
