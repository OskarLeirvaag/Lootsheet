# Contributing to LootSheet

Thank you for contributing to LootSheet.

## Prerequisites

- Go 1.26.1 or later
- `golangci-lint`
- `goimports`
- `govulncheck`

## Build

```sh
go build -o lootsheet .
```

## Test

```sh
go test ./...
```

## Quality Gate

Run the full local gate before opening a pull request:

```sh
make check
```

Pull requests should not be submitted unless `make check` passes.

## Pull Request Process

1. Fork the repository.
2. Create a branch from `main`.
3. Make your changes.
4. Run `make check`.
5. Submit a pull request against `main`.

## Project Conventions

Read these before making product or architecture changes:

- [README.md](README.md)
- [AGENTS.md](AGENTS.md)

When contributing, preserve the repository structure, dependency flow, and accounting invariants documented there.
