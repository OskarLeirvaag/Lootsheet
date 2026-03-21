.PHONY: build fmt imports test vet lint vuln deadcode manpages proto check

APP := lootsheet

build:
	go build -o $(APP) .
	GOOS=linux GOARCH=arm64 go build -o $(APP)-raspi .

fmt:
	@gofmt -l $$(find . -name '*.go' ! -name '*.pb.go' ! -path './.claude/*')

imports:
	@goimports -l $$(find . -name '*.go' ! -name '*.pb.go' ! -path './.claude/*')

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

vuln:
	govulncheck ./...

deadcode:
	deadcode $$(go list ./... | grep -v /testutil)

manpages:
	go run ./scripts/generate-manpages.go

proto:
	protoc --go_out=. --go_opt=paths=source_relative src/net/proto/lootsheet.proto

check: fmt imports test vet lint vuln
