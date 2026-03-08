.PHONY: build fmt test vet lint check

APP := lootsheet

build:
	go build -o $(APP) .

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

check: fmt test vet lint
