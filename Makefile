.PHONY: test build web clean

test:
	go test ./... -v

build:
	go build -o bin/issue2md ./cmd/issue2md

web:
	go build -o bin/issue2mdweb ./cmd/issue2mdweb

clean:
	rm -rf bin/
