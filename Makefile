.PHONY: build test lint docker-build clean

BINARY_CLI  := issue2md-cli
BINARY_WEB  := issue2md-web
BUILD_DIR   := bin
IMAGE_TAG   := issue2md:latest

build:
	go build -o $(BUILD_DIR)/$(BINARY_CLI) ./cmd/issue2md
	go build -o $(BUILD_DIR)/$(BINARY_WEB)  ./cmd/issue2mdweb

test:
	go test ./... -v

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE_TAG) .

clean:
	rm -rf $(BUILD_DIR)
