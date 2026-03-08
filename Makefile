.PHONY: build fmt imports test vet lint vuln deadcode manpages check

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

manpages:
	go run ./scripts/generate-manpages.go

check: fmt imports test vet lint vuln
