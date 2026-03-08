.PHONY: build fmt imports test vet lint vuln deadcode check

APP := lootsheet

build:
	go build -o $(APP) .

fmt:
	gofmt -l .

imports:
	goimports -l .

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

vuln:
	govulncheck ./...

deadcode:
	deadcode ./...

check: fmt imports test vet lint vuln
