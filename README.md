# go-isolate

[![CI](https://github.com/NemCaBong/go-isolate/actions/workflows/ci.yml/badge.svg)](https://github.com/NemCaBong/go-isolate/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/NemCaBong/go-isolate.svg)](https://pkg.go.dev/github.com/NemCaBong/go-isolate)
[![Go Report Card](https://goreportcard.com/badge/github.com/NemCaBong/go-isolate)](https://goreportcard.com/report/github.com/NemCaBong/go-isolate)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue)](go.mod)

A Go library for building and executing [isolate](https://github.com/ioi/isolate) sandbox commands with a **type-safe, fluent builder API**.

[Isolate](https://github.com/ioi/isolate) is the sandbox used by [Codeforces](https://codeforces.com), [CMS](https://cms-dev.github.io/), and the [IOI](https://ioinformatics.org/) to safely run untrusted code inside Linux containers with strict resource limits. This library wraps its CLI into an idiomatic Go interface — no manual string building, no missed flags.

## Why go-isolate?

| Without go-isolate | With go-isolate |
|--------------------|-----------------|
| Build shell strings by hand | Fluent builder with compile-time safety |
| Silently forget a flag | Validation catches config conflicts before execution |
| Parse meta-file output manually | `ParseMeta()` gives you a typed `Meta` struct |
| Manage init/run/cleanup yourself | `InitAndRun` handles the full lifecycle |

## Use Cases

- **Online judges / competitive programming platforms** — run user-submitted solutions with CPU/memory limits
- **Automated grading systems** — execute and evaluate programs in isolated environments
- **Security sandboxing** — run untrusted binaries without risking the host system
- **Batch code evaluation** — evaluate multiple submissions against test cases at scale

## Installation

```bash
go get github.com/NemCaBong/go-isolate
```

Requires the `isolate` binary to be installed on the host. See [isolate's installation guide](https://github.com/ioi/isolate#installation).

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    isolate "github.com/NemCaBong/go-isolate"
)

func main() {
    builder := isolate.New().
        BoxID(0).
        MemoryLimit(256 * 1024). // 256 MB
        TimeLimit(5.0).          // 5 seconds CPU
        WallTimeLimit(10.0).     // 10 seconds wall clock
        Meta("/tmp/meta.txt").
        Stdout("output.txt").
        Stderr("error.txt")

    if err := builder.Validate(); err != nil {
        log.Fatal(err)
    }

    exec := builder.Exec()
    ctx := context.Background()

    workDir, err := exec.Init(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Sandbox working directory:", workDir)

    // Copy your compiled binary into the sandbox before running.
    bin, _ := os.ReadFile("/path/to/compiled/solution")
    if err := exec.WriteToSandbox("solution", bin, 0755); err != nil {
        log.Fatal(err)
    }

    result, err := exec.Run(ctx, "./solution")
    if err != nil {
        log.Fatal(err)
    }

    if result.Meta != nil && result.Meta.IsSuccess() {
        fmt.Printf("Accepted! Time: %.3fs, Memory: %d KB\n",
            result.Meta.Time, result.Meta.MaxRSS)
    }

    exec.Cleanup(ctx)
}
```

## Features

### Fluent Builder

All 40+ isolate options are configured via method chaining:

```go
builder := isolate.New().
    BoxID(1).
    MemoryLimit(128 * 1024).
    TimeLimit(2.0).
    WallTimeLimit(5.0).
    Processes(1).
    ControlGroup().
    CGMemoryLimit(128 * 1024).
    InheritEnv("PATH").
    Dir("/usr", "/usr", isolate.DirMaybe)
```

### Command Building Without Execution

Inspect the exact command that would be run, without executing it:

```go
fmt.Println(builder.BuildInit().String())
// isolate --box-id=0 --init

fmt.Println(builder.BuildRun("./solution", "arg1").String())
// isolate --box-id=0 --mem=131072 --time=2 --wall-time=5 --run -- ./solution arg1

fmt.Println(builder.BuildCleanup().String())
// isolate --box-id=0 --cleanup
```

### Reusable Sandboxes

Use `InitAndRun` to run multiple programs against the same sandbox (reset between runs):

```go
exec := isolate.New().BoxID(0).MemoryLimit(256*1024).TimeLimit(5).Exec()

for _, tc := range testCases {
    result, err := exec.InitAndRun(ctx,
        func(workDir string) error {
            // Called once per run, after Init, before execution.
            // Copy input files and the binary here.
            return exec.WriteToSandbox("input.txt", tc.Input, 0644)
        },
        "./solution",
    )
    // result contains exit code, stdout, stderr, and parsed Meta
}
```

### Validation

Catch configuration errors before hitting the sandbox:

```go
builder := isolate.New().
    Stderr("error.txt").
    StderrToStdout() // mutually exclusive!

if err := builder.Validate(); err != nil {
    // Error: cannot use both --stderr and --stderr-to-stdout
}
```

### Meta-file Parsing

Parse isolate's structured output into a typed Go struct:

```go
meta, err := isolate.ParseMetaString(rawMetaOutput)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Status:  %s\n", meta.Status)   // OK, RE, SG, TO, XX
fmt.Printf("Time:    %.3fs\n", meta.Time)
fmt.Printf("Memory:  %d KB\n", meta.MaxRSS)
fmt.Printf("Exit:    %d\n", meta.ExitCode)
fmt.Printf("Success: %v\n", meta.IsSuccess())
```

### Dynamic Execute Options

Modify builder settings at execution time without rebuilding:

```go
exec.ApplyOptions(
    isolate.WithMemoryLimit(512 * 1024),
    isolate.WithTimeLimit(10.0),
)
```

## API Reference

### Basic Options

| Method | Flag | Description |
|--------|------|-------------|
| `BoxID(id)` | `--box-id` | Sandbox ID (default: 0) |
| `Meta(file)` | `--meta` | Meta-file output path |
| `Stdin(file)` | `--stdin` | Redirect stdin |
| `Stdout(file)` | `--stdout` | Redirect stdout |
| `Stderr(file)` | `--stderr` | Redirect stderr |
| `StderrToStdout()` | `--stderr-to-stdout` | Merge stderr into stdout |
| `Chdir(dir)` | `--chdir` | Change working directory inside sandbox |
| `Verbose()` | `--verbose` | Increase verbosity |
| `Silent()` | `--silent` | Suppress status messages |
| `Wait()` | `--wait` | Wait for other instances using same box |

### Resource Limits

| Method | Flag | Description |
|--------|------|-------------|
| `MemoryLimit(kb)` | `--mem` | Address space limit (KB) |
| `TimeLimit(sec)` | `--time` | CPU time limit (seconds) |
| `WallTimeLimit(sec)` | `--wall-time` | Wall clock limit (seconds) |
| `ExtraTime(sec)` | `--extra-time` | Grace period after timeout |
| `StackLimit(kb)` | `--stack` | Stack size limit (KB) |
| `OpenFilesLimit(n)` | `--open-files` | Max open files (0 = unlimited) |
| `FileSizeLimit(kb)` | `--fsize` | Max output file size (KB) |
| `DiskQuota(blocks, inodes)` | `--quota` | Disk quota (init only) |
| `CoreLimit(kb)` | `--core` | Core dump size limit (KB) |
| `Processes(n)` | `--processes` | Max simultaneous processes/threads |
| `ProcessesUnlimited()` | `--processes` | No process count limit |

### Environment Rules

| Method | Flag | Description |
|--------|------|-------------|
| `FullEnv()` | `--full-env` | Inherit all environment variables |
| `InheritEnv(var)` | `--env=var` | Inherit a specific variable |
| `SetEnv(var, val)` | `--env=var=val` | Set a variable to a value |
| `RemoveEnv(var)` | `--env=var=` | Explicitly remove a variable |

### Directory Rules

| Method | Flag | Description |
|--------|------|-------------|
| `NoDefaultDirs()` | `--no-default-dirs` | Disable default directory bindings |
| `Dir(in, out, opts...)` | `--dir=in=out:opts` | Bind host dir to sandbox path |
| `DirSimple(dir, opts...)` | `--dir=dir:opts` | Bind with same path inside/outside |
| `RemoveDir(in)` | `--dir=in=` | Remove a directory rule |

**Directory options:** `DirRW` `DirDev` `DirNoExec` `DirMaybe` `DirFS` `DirTmp` `DirNoRec`

### Control Groups

| Method | Flag | Description |
|--------|------|-------------|
| `ControlGroup()` | `--cg` | Enable cgroup-based resource accounting |
| `CGMemoryLimit(kb)` | `--cg-mem` | Memory limit enforced via cgroups (KB) |

### Special Options

| Method | Flag | Description |
|--------|------|-------------|
| `ShareNet()` | `--share-net` | Share host network namespace |
| `InheritFDs()` | `--inherit-fds` | Keep parent file descriptors |
| `TTYHack()` | `--tty-hack` | Handle programs that require a TTY |
| `SpecialFiles()` | `--special-files` | Keep special files accessible |
| `AsUID(uid)` | `--as-uid` | Run as a specific UID |
| `AsGID(gid)` | `--as-gid` | Run as a specific GID |

### Meta-file Status Codes

| Constant | Value | Meaning |
|----------|-------|---------|
| `StatusOK` | `""` | Program terminated normally |
| `StatusRuntimeError` | `RE` | Non-zero exit code |
| `StatusSignal` | `SG` | Killed by a signal |
| `StatusTimeout` | `TO` | Time limit exceeded |
| `StatusInternalError` | `XX` | Internal sandbox error |

## Project Structure

```
go-isolate/
├── isolate.go          # Core types, constants, and DirOption definitions
├── builder.go          # Fluent builder (40+ methods)
├── command.go          # Command construction: BuildInit, BuildRun, BuildCleanup
├── executor.go         # Sandbox lifecycle: Init, Run, Cleanup, InitAndRun
├── execute_option.go   # Functional options for dynamic builder modification
├── meta.go             # Meta-file parser → typed Meta struct
├── validate.go         # Configuration validation (10+ checks)
├── *_test.go           # Unit tests
└── example/main.go     # Runnable examples
```

## Contributing

Contributions are welcome! Please open an issue before submitting a large PR so we can discuss the approach.

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Add tests for your changes
4. Run `go test ./...` and `go vet ./...`
5. Open a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

## Related Projects

- [isolate](https://github.com/ioi/isolate) — the underlying sandbox tool
- [go-sandbox](https://github.com/criyle/go-sandbox) — alternative Go sandbox using seccomp/cgroups directly
- [CMS](https://cms-dev.github.io/) — contest management system that uses isolate

## License

[MIT](LICENSE) — Copyright (c) 2024 Hoang Nguyen Minh