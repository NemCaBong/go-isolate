# Contributing to go-isolate

Thank you for your interest in contributing!

## Getting started

```bash
git clone https://github.com/NemCaBong/go-isolate.git
cd go-isolate
go mod download
go test ./...
```

## Guidelines

- **Open an issue first** for any significant change so we can align on the approach before you invest time writing code.
- **Keep PRs focused** — one feature or fix per PR.
- **Add tests** for any new builder options, validation rules, or meta-parsing behavior.
- **No external dependencies** — this library intentionally uses only the Go standard library.

## Running tests

```bash
go test -v -race ./...
go vet ./...
```

Note: tests that exercise `executor.go` require `isolate` to be installed. Builder, validation, and meta-parsing tests work without it.

## Reporting bugs

Please use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.yml) and include a minimal reproducer.