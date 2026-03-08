package isolate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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

	// For graceful shutdown
	cleanupOnce sync.Once
	stopSignal  chan os.Signal
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
	defer dst.Close()

	if _, err := io.Copy(dst, r); err != nil {
		return fmt.Errorf("failed to write file %s: %w", destPath, err)
	}

	return nil
}

// --- Manual Lifecycle ---

// Init initializes the sandbox and returns the working directory path.
// If the sandbox already exists, it is reset.
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
// It is safe to call Cleanup multiple times.
func (e *Executor) Cleanup(ctx context.Context) error {
	var cleanupErr error
	e.cleanupOnce.Do(func() {
		cmd := e.builder.BuildCleanup()
		result, err := e.execute(ctx, cmd)
		if err != nil {
			cleanupErr = fmt.Errorf("cleanup failed: %w", err)
			return
		}

		if result.ExitCode != 0 {
			cleanupErr = fmt.Errorf("cleanup failed with exit code %d: %s", result.ExitCode, result.Stderr)
			return
		}

		e.workDir = ""
	})
	return cleanupErr
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
	// Reset cleanupOnce so Cleanup works again after re-init
	e.cleanupOnce = sync.Once{}

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

// --- Graceful Shutdown ---

// EnableAutoCleanup registers OS signal handlers (SIGINT, SIGTERM) so that
// the sandbox is automatically cleaned up when the application is shutting down.
//
// This handles the case where your app receives Ctrl+C or a kill signal.
// Note: SIGKILL (kill -9) cannot be caught — for that case, the sandbox
// will be reset automatically on the next Init call.
//
// Call DisableAutoCleanup or Cleanup to stop listening for signals.
//
// Usage:
//
//	exec := isolate.New().BoxID(0).Exec()
//	exec.Init(ctx)
//	exec.EnableAutoCleanup()  // auto-cleanup on SIGINT/SIGTERM
//	defer exec.Cleanup(ctx)   // also cleanup normally
func (e *Executor) EnableAutoCleanup() {
	e.stopSignal = make(chan os.Signal, 1)
	signal.Notify(e.stopSignal, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig, ok := <-e.stopSignal
		if !ok {
			return // channel closed, normal shutdown
		}

		fmt.Fprintf(os.Stderr, "\n[go-isolate] Received %s, cleaning up sandbox...\n", sig)

		// Use a background context since the original may be cancelled
		ctx, cancel := context.WithTimeout(context.Background(), 5*0x1) // 5 seconds
		defer cancel()

		if err := e.Cleanup(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "[go-isolate] Cleanup error: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "[go-isolate] Sandbox cleaned up successfully\n")
		}

		// Re-raise the signal so the process exits with the correct status
		signal.Reset(sig)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(sig)
	}()
}

// DisableAutoCleanup stops listening for OS signals.
func (e *Executor) DisableAutoCleanup() {
	if e.stopSignal != nil {
		signal.Stop(e.stopSignal)
		close(e.stopSignal)
		e.stopSignal = nil
	}
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
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			return nil, err
		}
	}

	return result, nil
}
