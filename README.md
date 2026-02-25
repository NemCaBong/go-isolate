# go-isolate

A Go library for building and executing [isolate](https://github.com/ioi/isolate) sandbox commands using the **builder design pattern**.

Isolate is a tool for running processes inside Linux sandboxes (containers) with strict resource limits and filesystem isolation. This library provides a **type-safe, fluent API** for constructing isolate commands programmatically.

## Installation

```bash
go get github.com/hoangnm/go-isolate
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    isolate "github.com/hoangnm/go-isolate"
)

func main() {
    // Create a builder with desired settings
    builder := isolate.New().
        BoxID(0).
        MemoryLimit(256 * 1024).  // 256 MB
        TimeLimit(5.0).           // 5 seconds
        WallTimeLimit(10.0).      // 10 seconds
        Meta("/tmp/meta.txt").
        Stdin("input.txt").
        Stdout("output.txt").
        Stderr("error.txt")

    // Validate configuration
    if err := builder.Validate(); err != nil {
        log.Fatal(err)
    }

    // Create an executor
    exec := builder.Exec()
    ctx := context.Background()

    // Initialize sandbox
    workDir, err := exec.Init(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Working directory:", workDir)

    // Run a program
    result, err := exec.Run(ctx, "./solution", "arg1", "arg2")
    if err != nil {
        log.Fatal(err)
    }

    // Check results
    if result.Meta != nil && result.Meta.IsSuccess() {
        fmt.Printf("Success! Time: %.3fs, Memory: %d KB\n",
            result.Meta.Time, result.Meta.MaxRSS)
    }

    // Clean up
    exec.Cleanup(ctx)
}
```

## Features

### Builder Pattern

All isolate options are configured via a fluent builder API:

```go
builder := isolate.New().
    BoxID(0).
    MemoryLimit(256 * 1024).
    TimeLimit(5.0).
    WallTimeLimit(10.0)
```

### Command Building (Without Execution)

If you only need to build the command-line arguments without executing them:

```go
// Build init command
initCmd := builder.BuildInit()
fmt.Println(initCmd.String())
// Output: isolate --box-id=0 --init

// Build run command
runCmd := builder.BuildRun("./solution", "arg1")
fmt.Println(runCmd.String())
// Output: isolate --box-id=0 --mem=262144 --time=5 --wall-time=10 --run -- ./solution arg1

// Build cleanup command
cleanupCmd := builder.BuildCleanup()
fmt.Println(cleanupCmd.String())
// Output: isolate --box-id=0 --cleanup
```

### Validation

The builder includes validation for common configuration errors:

```go
builder := isolate.New().
    Stderr("error.txt").
    StderrToStdout()  // Mutually exclusive with Stderr!

if err := builder.Validate(); err != nil {
    // Error: cannot use both Stderr and StderrToStdout
}
```

### Meta-file Parsing

Parse isolate's meta-file output into structured data:

```go
meta, err := isolate.ParseMetaString(metaData)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Time: %.3fs\n", meta.Time)
fmt.Printf("Memory: %d KB\n", meta.MaxRSS)
fmt.Printf("Status: %s\n", meta.Status)
```

## API Reference

### Basic Options

| Method | Flag | Description |
|--------|------|-------------|
| `BoxID(id)` | `--box-id` | Set sandbox ID (default: 0) |
| `Meta(file)` | `--meta` | Output meta-file path |
| `Stdin(file)` | `--stdin` | Redirect stdin |
| `Stdout(file)` | `--stdout` | Redirect stdout |
| `Stderr(file)` | `--stderr` | Redirect stderr |
| `StderrToStdout()` | `--stderr-to-stdout` | Redirect stderr to stdout |
| `Chdir(dir)` | `--chdir` | Change working directory |
| `Verbose()` | `--verbose` | Increase verbosity |
| `Silent()` | `--silent` | Suppress status messages |
| `Wait()` | `--wait` | Wait for other instances |

### Limits

| Method | Flag | Description |
|--------|------|-------------|
| `MemoryLimit(kb)` | `--mem` | Address space limit (KB) |
| `TimeLimit(sec)` | `--time` | CPU time limit (seconds) |
| `WallTimeLimit(sec)` | `--wall-time` | Wall clock limit (seconds) |
| `ExtraTime(sec)` | `--extra-time` | Grace period after timeout |
| `StackLimit(kb)` | `--stack` | Stack size limit (KB) |
| `OpenFilesLimit(n)` | `--open-files` | Max open files (0=unlimited) |
| `FileSizeLimit(kb)` | `--fsize` | Max file size (KB) |
| `DiskQuota(blocks, inodes)` | `--quota` | Disk quota (init only) |
| `CoreLimit(kb)` | `--core` | Core file size limit (KB) |
| `Processes(n)` | `--processes` | Max processes/threads |
| `ProcessesUnlimited()` | `--processes` | Unlimited processes |

### Environment Rules

| Method | Flag | Description |
|--------|------|-------------|
| `FullEnv()` | `--full-env` | Inherit all env vars |
| `InheritEnv(var)` | `--env=var` | Inherit specific var |
| `SetEnv(var, val)` | `--env=var=val` | Set env var |
| `RemoveEnv(var)` | `--env=var=` | Remove env var |

### Directory Rules

| Method | Flag | Description |
|--------|------|-------------|
| `NoDefaultDirs()` | `--no-default-dirs` | Disable default bindings |
| `Dir(in, out, opts...)` | `--dir=in=out:opts` | Bind directory |
| `DirSimple(dir, opts...)` | `--dir=dir:opts` | Bind with same path |
| `RemoveDir(in)` | `--dir=in=` | Remove directory rule |

#### Directory Options

- `DirRW` — Read-write access
- `DirDev` — Allow devices
- `DirNoExec` — No execution
- `DirMaybe` — Ignore if missing
- `DirFS` — Mount filesystem
- `DirTmp` — Temporary directory
- `DirNoRec` — Non-recursive bind

### Control Groups

| Method | Flag | Description |
|--------|------|-------------|
| `ControlGroup()` | `--cg` | Enable control groups |
| `CGMemoryLimit(kb)` | `--cg-mem` | CG memory limit (KB) |

### Special Options

| Method | Flag | Description |
|--------|------|-------------|
| `ShareNet()` | `--share-net` | Share network namespace |
| `InheritFDs()` | `--inherit-fds` | Keep parent FDs |
| `TTYHack()` | `--tty-hack` | Handle interactive TTY |
| `SpecialFiles()` | `--special-files` | Keep special files |
| `AsUID(uid)` | `--as-uid` | Act as user |
| `AsGID(gid)` | `--as-gid` | Act as group |

### Meta-file Status Codes

| Constant | Code | Description |
|----------|------|-------------|
| `StatusOK` | (empty) | Normal termination |
| `StatusRuntimeError` | `RE` | Non-zero exit code |
| `StatusSignal` | `SG` | Killed by signal |
| `StatusTimeout` | `TO` | Time limit exceeded |
| `StatusInternalError` | `XX` | Internal sandbox error |

## Project Structure

```
go-isolate/
├── isolate.go         # Core types and constants
├── builder.go         # Builder design pattern implementation
├── command.go         # Command building logic
├── executor.go        # Command execution with context support
├── meta.go            # Meta-file parsing
├── validate.go        # Configuration validation
├── builder_test.go    # Builder tests
├── meta_test.go       # Meta-file parsing tests
├── validate_test.go   # Validation tests
├── example/
│   └── main.go        # Usage examples
├── go.mod
└── README.md
```

## License

MIT License
