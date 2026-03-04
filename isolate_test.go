package isolate_test

import (
	"testing"

	isolate "github.com/NemCaBong/go-isolate"
)

func TestInitAndRunWorkflow(t *testing.T) {
	// Verify the builder produces correct commands for the reusable pattern:
	// InitAndRun → InitAndRun → Cleanup (only at the end)
	builder := isolate.New().
		BoxID(1).
		MemoryLimit(128 * 1024).
		TimeLimit(2.0).
		WallTimeLimit(5.0).
		Meta("/tmp/meta.txt")

	// First init (what InitAndRun does internally)
	init1 := builder.BuildInit()
	assertContains(t, init1.Args, "--box-id=1")
	assertContains(t, init1.Args, "--init")

	// First run
	run1 := builder.BuildRun("./solution1")
	assertContains(t, run1.Args, "--mem=131072")
	assertContains(t, run1.Args, "--time=2")
	assertContains(t, run1.Args, "--wall-time=5")
	assertContains(t, run1.Args, "--run")
	assertContains(t, run1.Args, "./solution1")

	// Second init (reset) — no cleanup needed between runs
	init2 := builder.BuildInit()
	assertContains(t, init2.Args, "--box-id=1")
	assertContains(t, init2.Args, "--init")

	// Second run
	run2 := builder.BuildRun("./solution2", "arg1")
	assertContains(t, run2.Args, "--run")
	assertContains(t, run2.Args, "./solution2")
	assertContains(t, run2.Args, "arg1")

	// Final cleanup — only once at the end
	cleanup := builder.BuildCleanup()
	assertContains(t, cleanup.Args, "--box-id=1")
	assertContains(t, cleanup.Args, "--cleanup")
}

func TestExecutorWorkDir(t *testing.T) {
	exec := isolate.New().BoxID(0).Exec()

	// WorkDir should be empty before Init
	if exec.WorkDir() != "" {
		t.Errorf("expected empty WorkDir before Init, got %q", exec.WorkDir())
	}
}
