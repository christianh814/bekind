.PHONY: all build run test clean

# Variables
BINARY_NAME=bekind
VERSION=$(git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"
GO_FILES=$(find . -name '*.go')

# Targets
all: build

build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/bekind

run: build
	./$(BINARY_NAME)

test:
	go test -race -cover ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
