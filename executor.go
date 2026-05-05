package isolate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PrepareFunc is a callback invoked between Init and Run, typically used
// to write files into the sandbox. The workDir parameter is the sandbox
// working directory (e.g., /var/local/lib/isolate/0/box).
type PrepareFunc func(workDir string) error

// Result holds the outcome of running a program via isolate.
type Result struct {
	// ExitCode is the exit code of the isolate process itself.
	// 0 = program finished correctly, 1 = program finished incorrectly,
	// other = internal sandbox error.
	ExitCode int
	// Stdout contains the standard output of the isolate process.
	Stdout string
	// Stderr contains the standard error of the isolate process.
	Stderr string
	// Meta contains the parsed meta-file data (if a meta file was configured).
	Meta *Meta
}

// Executor provides methods for running isolate commands.
// It is the single entry point for all isolate operations — both manual
// lifecycle management and the reusable sandbox pattern.
type Executor struct {
	builder *Builder
	workDir string
}

// NewExecutor creates a new Executor from a Builder.
func NewExecutor(b *Builder) *Executor {
	return &Executor{builder: b}
}

// Exec creates an Executor from the builder for convenience.
func (b *Builder) Exec() *Executor {
	return NewExecutor(b)
}

// WorkDir returns the working directory of the sandbox after Init has been called.
// Returns empty string if Init has not been called yet.
func (e *Executor) WorkDir() string {
	return e.workDir
}

// --- File Operations ---

// WriteToSandbox writes content into a file inside the sandbox.
// destName is the filename (or relative path) inside the sandbox.
// The caller is responsible for reading source files; this method only
// receives the content to write.
//
// Example:
//
//	exec.Init(ctx)
//
//	// Write a compiled binary (read by caller)
//	bin, _ := os.ReadFile("/path/to/compiled/solution")
//	exec.WriteToSandbox("solution", bin, 0755)
//
//	// Write input data
//	exec.WriteToSandbox("input.txt", []byte("5\n1 2 3 4 5\n"), 0644)
//
//	exec.Run(ctx, "./solution")
func (e *Executor) WriteToSandbox(destName string, content []byte, perm os.FileMode) error {
	if e.workDir == "" {
		return fmt.Errorf("sandbox not initialized: call Init() first")
	}

	destPath := filepath.Join(e.workDir, destName)

	// Ensure parent directories exist
	if dir := filepath.Dir(destPath); dir != e.workDir {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	if err := os.WriteFile(destPath, content, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// WriteReaderToSandbox writes content from an io.Reader into a file inside
// the sandbox. This is useful for large files where loading the entire content
// into memory is not desirable (e.g., streaming from HTTP response, database, etc.).
//
// Example:
//
//	// Stream from an HTTP response
//	resp, _ := http.Get("https://example.com/solution")
//	defer resp.Body.Close()
//	exec.WriteReaderToSandbox("solution", resp.Body, 0755)
//
//	// Stream from an open file
//	f, _ := os.Open("/path/to/large/binary")
//	defer f.Close()
//	exec.WriteReaderToSandbox("binary", f, 0755)
func (e *Executor) WriteReaderToSandbox(destName string, r io.Reader, perm os.FileMode) error {
	if e.workDir == "" {
		return fmt.Errorf("sandbox not initialized: call Init() first")
	}

	destPath := filepath.Join(e.workDir, destName)

	// Ensure parent directories exist
	if dir := filepath.Dir(destPath); dir != e.workDir {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}

	if _, copyErr := io.Copy(dst, r); copyErr != nil {
		// Best-effort close; the copy error is the meaningful one to return.
		_ = dst.Close()
		return fmt.Errorf("failed to write file %s: %w", destPath, copyErr)
	}

	// Close explicitly (not via defer) so flush errors are not silently dropped.
	if closeErr := dst.Close(); closeErr != nil {
		return fmt.Errorf("failed to close file %s: %w", destPath, closeErr)
	}

	return nil
}

// --- Manual Lifecycle ---

// Init initializes the sandbox and returns the working directory path.
// If the sandbox already exists, it is reset.
// Per isolate docs: --init always exits 0 even when resetting an existing sandbox.
func (e *Executor) Init(ctx context.Context) (string, error) {
	cmd := e.builder.BuildInit()
	result, err := e.execute(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("init failed: %w", err)
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("init failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	e.workDir = strings.TrimSpace(result.Stdout)
	return e.workDir, nil
}

// Run executes a program inside the sandbox.
//
// NOTE: You must ensure the program/binary and any required input files
// are already present inside the sandbox before calling Run. Use
// WriteToSandbox or WriteReaderToSandbox to copy files after Init.
func (e *Executor) Run(ctx context.Context, program string, args ...string) (*Result, error) {
	cmd := e.builder.BuildRun(program, args...)
	result, err := e.execute(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("run failed: %w", err)
	}

	// Attempt to parse meta-file if configured
	if e.builder.meta != "" {
		metaData, readErr := os.ReadFile(e.builder.meta)
		if readErr == nil {
			meta, parseErr := ParseMetaString(string(metaData))
			if parseErr == nil {
				result.Meta = meta
			}
		}
	}

	return result, nil
}

// Cleanup removes the sandbox and its temporary files.
// Safe to call multiple times — per isolate docs, --cleanup is idempotent:
// if the sandbox was already removed or never initialized, it does nothing
// and exits with code 0. Errors from the isolate process are therefore ignored.
func (e *Executor) Cleanup(ctx context.Context) {
	cmd := e.builder.BuildCleanup()
	// Discard result: --cleanup always exits 0 per isolate docs.
	_, _ = e.execute(ctx, cmd)
	e.workDir = ""
}

// --- Reusable Sandbox Pattern ---

// InitAndRun re-initializes the sandbox (resetting it), runs an optional
// prepare function to write files into the sandbox, and then runs the program.
// This is the reusable sandbox pattern — no need to call Cleanup between runs,
// because Init resets the sandbox automatically.
//
// The optional prepare function is called after Init with the sandbox working
// directory path, allowing you to write files into the sandbox before execution.
//
// Usage:
//
//	exec := isolate.New().BoxID(0).MemoryLimit(256*1024).TimeLimit(5).Exec()
//	defer exec.Cleanup(ctx)
//
//	// Simple: no file preparation needed
//	result1, _ := exec.InitAndRun(ctx, nil, "./solution1")
//
//	// With preparation: write files before running
//	bin, _ := os.ReadFile("/host/solution")
//	prepare := func(workDir string) error {
//		return exec.WriteToSandbox("solution", bin, 0755)
//	}
//	result2, _ := exec.InitAndRun(ctx, prepare, "./solution")
func (e *Executor) InitAndRun(ctx context.Context, prepare PrepareFunc, program string, args ...string) (*Result, error) {
	// Per isolate docs: --init resets an existing sandbox automatically,
	// so no --cleanup is needed between runs.
	workDir, err := e.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("sandbox reset failed: %w", err)
	}

	// Run prepare function if provided (e.g., copy files into sandbox)
	if prepare != nil {
		if err := prepare(workDir); err != nil {
			return nil, fmt.Errorf("prepare failed: %w", err)
		}
	}

	return e.Run(ctx, program, args...)
}

// --- Control Group ---

// PrintCGRoot prints the control group root path.
func (e *Executor) PrintCGRoot(ctx context.Context) (string, error) {
	cmd := e.builder.BuildPrintCGRoot()
	result, err := e.execute(ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("print-cg-root failed: %w", err)
	}

	return strings.TrimSpace(result.Stdout), nil
}

// execute runs an isolate command and captures output.
func (e *Executor) execute(ctx context.Context, cmd *Command) (*Result, error) {
	execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args...)

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	err := execCmd.Run()

	result := &Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if e.builder.stdout != "" && e.workDir != "" {
		data, readErr := os.ReadFile(filepath.Join(e.workDir, e.builder.stdout))
		if readErr == nil {
			result.Stdout = string(data)
		}
	}
	if e.builder.stderr != "" && e.workDir != "" {
		data, readErr := os.ReadFile(filepath.Join(e.workDir, e.builder.stderr))
		if readErr == nil {
			result.Stderr = string(data)
		}
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	return result, nil
}

func (e *Executor) ApplyOptions(options ...ExecuteOption) {
	for _, opt := range options {
		opt(e)
	}
}
