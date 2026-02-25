// Example demonstrates usage of the go-isolate library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	isolate "github.com/hoangnm/go-isolate"
)

func main() {
	fmt.Println("=== Example 1: Build Commands Only ===")
	buildExample()

	fmt.Println("\n=== Example 2: Full Lifecycle (Init → Write → Run → Cleanup) ===")
	fullLifecycleExample()

	fmt.Println("\n=== Example 3: Reusable Pattern (InitAndRun) ===")
	reusableExample()

	fmt.Println("\n=== Example 4: Auto Cleanup on Crash/Signal ===")
	autoCleanupExample()

	fmt.Println("\n=== Example 5: Meta-file Parsing ===")
	metaExample()
}

// buildExample shows how to build commands without executing them.
func buildExample() {
	builder := isolate.New().
		BoxID(0).
		MemoryLimit(256 * 1024).
		TimeLimit(5.0).
		WallTimeLimit(10.0).
		Meta("/tmp/meta.txt").
		Stdin("input.txt").
		Stdout("output.txt").
		Stderr("error.txt")

	fmt.Printf("Init:    %s\n", builder.BuildInit().String())
	fmt.Printf("Run:     %s\n", builder.BuildRun("./solution").String())
	fmt.Printf("Cleanup: %s\n", builder.BuildCleanup().String())
}

// fullLifecycleExample shows the complete workflow:
// 1. Init sandbox → get workDir
// 2. Write executable + input content into the sandbox
// 3. Run the program
// 4. Cleanup
func fullLifecycleExample() {
	exec := isolate.New().
		BoxID(0).
		MemoryLimit(256 * 1024).
		TimeLimit(5.0).
		WallTimeLimit(10.0).
		Meta("/tmp/meta.txt").
		Stdin("input.txt").
		Stdout("output.txt").
		Exec()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer exec.Cleanup(ctx)

	// Step 1: Init — creates sandbox and returns the working directory
	workDir, err := exec.Init(ctx)
	if err != nil {
		log.Printf("Note: Requires isolate to be installed: %v", err)
		return
	}
	fmt.Printf("Sandbox path: %s\n", workDir)

	// Step 2: Write files into the sandbox
	// Read the compiled binary on the host, then write its content into sandbox
	bin, err := os.ReadFile("/path/to/compiled/solution")
	if err != nil {
		log.Fatalf("Failed to read solution binary: %v", err)
	}
	if err := exec.WriteToSandbox("solution", bin, 0755); err != nil {
		log.Fatalf("Failed to write solution to sandbox: %v", err)
	}

	// Write input data directly
	if err := exec.WriteToSandbox("input.txt", []byte("5\n1 2 3 4 5\n"), 0644); err != nil {
		log.Fatalf("Failed to write input: %v", err)
	}

	// Step 3: Run the program inside the sandbox
	result, err := exec.Run(ctx, "./solution")
	if err != nil {
		log.Fatalf("Run failed: %v", err)
	}
	fmt.Printf("Exit code: %d\n", result.ExitCode)
	fmt.Printf("Stdout: %s\n", result.Stdout)
}

// reusableExample shows the reusable sandbox pattern — no cleanup between runs.
// Uses PrepareFunc to write files before each run.
func reusableExample() {
	exec := isolate.New().
		BoxID(0).
		MemoryLimit(256 * 1024).
		TimeLimit(5.0).
		WallTimeLimit(10.0).
		Meta("/tmp/meta.txt").
		Stdin("input.txt").
		Stdout("output.txt").
		Exec()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Only cleanup when completely done
	defer exec.Cleanup(ctx)

	// Each solution binary path to test
	solutionPaths := []string{"/host/solution1", "/host/solution2", "/host/solution3"}

	for i, path := range solutionPaths {
		// Read the binary content on the host
		bin, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Run %d: failed to read binary: %v", i+1, err)
			continue
		}

		// PrepareFunc writes the content into sandbox between Init and Run
		prepare := func(workDir string) error {
			// Write the solution binary content
			if err := exec.WriteToSandbox("solution", bin, 0755); err != nil {
				return err
			}
			// Write test input content
			return exec.WriteToSandbox("input.txt", []byte("test input\n"), 0644)
		}

		result, err := exec.InitAndRun(ctx, prepare, "./solution")
		if err != nil {
			log.Printf("Run %d failed: %v", i+1, err)
			continue
		}
		fmt.Printf("Run %d: exit=%d\n", i+1, result.ExitCode)
	}
}

// autoCleanupExample shows automatic cleanup when the app receives a signal.
func autoCleanupExample() {
	exec := isolate.New().
		BoxID(0).
		MemoryLimit(256 * 1024).
		TimeLimit(5.0).
		Exec()

	ctx := context.Background()

	// Step 1: Initialize the sandbox
	_, err := exec.Init(ctx)
	if err != nil {
		log.Printf("Note: Requires isolate to be installed: %v", err)
		return
	}

	// Enable auto-cleanup: if the app receives SIGINT (Ctrl+C) or SIGTERM,
	// the sandbox is automatically cleaned up before the process exits.
	exec.EnableAutoCleanup()
	defer exec.Cleanup(ctx) // also cleanup on normal exit

	// Step 2: Run the program (using system binary already available in sandbox)
	result, err := exec.Run(ctx, "/bin/echo", "Hello!")
	if err != nil {
		log.Printf("Run failed: %v", err)
		return
	}
	fmt.Printf("Exit code: %d, Stdout: %s", result.ExitCode, result.Stdout)
}

// metaExample shows how to parse isolate meta-file output.
func metaExample() {
	data := `time:1.234
time-wall:2.345
max-rss:12345
csw-forced:10
csw-voluntary:20
exitcode:0
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		log.Fatalf("Failed to parse meta: %v", err)
	}

	fmt.Printf("CPU Time:     %.3f seconds\n", meta.Time)
	fmt.Printf("Wall Time:    %.3f seconds\n", meta.TimeWall)
	fmt.Printf("Memory (RSS): %d KB\n", meta.MaxRSS)
	fmt.Printf("Exit Code:    %d\n", meta.ExitCode)
	fmt.Printf("Success:      %v\n", meta.IsSuccess())
}
