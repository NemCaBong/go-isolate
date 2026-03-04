package isolate_test

import (
	"testing"

	isolate "github.com/NemCaBong/go-isolate"
)

func TestBuilderBuildInit(t *testing.T) {
	cmd := isolate.New().
		BoxID(1).
		BuildInit()

	assertContains(t, cmd.Args, "--box-id=1")
	assertContains(t, cmd.Args, "--init")
}

func TestBuilderBuildInitWithQuota(t *testing.T) {
	cmd := isolate.New().
		BoxID(0).
		DiskQuota(1000, 500).
		BuildInit()

	assertContains(t, cmd.Args, "--quota=1000,500")
	assertContains(t, cmd.Args, "--init")
}

func TestBuilderBuildRun(t *testing.T) {
	cmd := isolate.New().
		BoxID(0).
		MemoryLimit(256000).
		TimeLimit(5.0).
		WallTimeLimit(10.0).
		ExtraTime(0.5).
		StackLimit(8192).
		OpenFilesLimit(128).
		FileSizeLimit(1024).
		CoreLimit(0).
		Processes(10).
		Meta("/tmp/meta.txt").
		Stdin("input.txt").
		Stdout("output.txt").
		Stderr("error.txt").
		Chdir("/box").
		Verbose().
		BuildRun("./solution", "arg1", "arg2")

	assertContains(t, cmd.Args, "--box-id=0")
	assertContains(t, cmd.Args, "--mem=256000")
	assertContains(t, cmd.Args, "--time=5")
	assertContains(t, cmd.Args, "--wall-time=10")
	assertContains(t, cmd.Args, "--extra-time=0.5")
	assertContains(t, cmd.Args, "--stack=8192")
	assertContains(t, cmd.Args, "--open-files=128")
	assertContains(t, cmd.Args, "--fsize=1024")
	assertContains(t, cmd.Args, "--core=0")
	assertContains(t, cmd.Args, "--processes=10")
	assertContains(t, cmd.Args, "--meta=/tmp/meta.txt")
	assertContains(t, cmd.Args, "--stdin=input.txt")
	assertContains(t, cmd.Args, "--stdout=output.txt")
	assertContains(t, cmd.Args, "--stderr=error.txt")
	assertContains(t, cmd.Args, "--chdir=/box")
	assertContains(t, cmd.Args, "--verbose")
	assertContains(t, cmd.Args, "--run")
	assertContains(t, cmd.Args, "--")
	assertContains(t, cmd.Args, "./solution")
	assertContains(t, cmd.Args, "arg1")
	assertContains(t, cmd.Args, "arg2")
}

func TestBuilderBuildCleanup(t *testing.T) {
	cmd := isolate.New().
		BoxID(2).
		BuildCleanup()

	assertContains(t, cmd.Args, "--box-id=2")
	assertContains(t, cmd.Args, "--cleanup")
}

func TestBuilderEnvironment(t *testing.T) {
	cmd := isolate.New().
		FullEnv().
		InheritEnv("PATH").
		SetEnv("HOME", "/home/user").
		RemoveEnv("DEBUG").
		BuildRun("./program")

	assertContains(t, cmd.Args, "--full-env")
	assertContains(t, cmd.Args, "--env=PATH")
	assertContains(t, cmd.Args, "--env=HOME=/home/user")
	assertContains(t, cmd.Args, "--env=DEBUG=")
}

func TestBuilderDirectoryRules(t *testing.T) {
	cmd := isolate.New().
		NoDefaultDirs().
		Dir("box", "/home/user/box", isolate.DirRW).
		DirSimple("proc", isolate.DirFS).
		Dir("tmp", "", isolate.DirTmp).
		RemoveDir("lib64").
		BuildRun("./program")

	assertContains(t, cmd.Args, "--no-default-dirs")
	assertContains(t, cmd.Args, "--dir=box=/home/user/box:rw")
	assertContains(t, cmd.Args, "--dir=proc:fs")
	assertContains(t, cmd.Args, "--dir=tmp:tmp")
	assertContains(t, cmd.Args, "--dir=lib64=")
}

func TestBuilderDirectoryMultipleOptions(t *testing.T) {
	cmd := isolate.New().
		Dir("data", "/mnt/data", isolate.DirRW, isolate.DirDev, isolate.DirNoRec).
		BuildRun("./program")

	assertContains(t, cmd.Args, "--dir=data=/mnt/data:rw,dev,norec")
}

func TestBuilderControlGroups(t *testing.T) {
	cmd := isolate.New().
		ControlGroup().
		CGMemoryLimit(512000).
		BuildRun("./program")

	assertContains(t, cmd.Args, "--cg")
	assertContains(t, cmd.Args, "--cg-mem=512000")
}

func TestBuilderSpecialOptions(t *testing.T) {
	cmd := isolate.New().
		ShareNet().
		InheritFDs().
		TTYHack().
		SpecialFiles().
		AsUID(1000).
		AsGID(1000).
		BuildRun("./program")

	assertContains(t, cmd.Args, "--share-net")
	assertContains(t, cmd.Args, "--inherit-fds")
	assertContains(t, cmd.Args, "--tty-hack")
	assertContains(t, cmd.Args, "--special-files")
	assertContains(t, cmd.Args, "--as-uid=1000")
	assertContains(t, cmd.Args, "--as-gid=1000")
}

func TestBuilderStderrToStdout(t *testing.T) {
	cmd := isolate.New().
		StderrToStdout().
		BuildRun("./program")

	assertContains(t, cmd.Args, "--stderr-to-stdout")
}

func TestBuilderProcessesUnlimited(t *testing.T) {
	cmd := isolate.New().
		ProcessesUnlimited().
		BuildRun("./program")

	assertContains(t, cmd.Args, "--processes")

	// Should not contain --processes= (with a value)
	for _, arg := range cmd.Args {
		if arg != "--processes" && len(arg) > len("--processes") && arg[:len("--processes=")] == "--processes=" {
			t.Errorf("expected --processes without value, got %q", arg)
		}
	}
}

func TestBuilderSilent(t *testing.T) {
	cmd := isolate.New().
		Silent().
		BuildRun("./program")

	assertContains(t, cmd.Args, "--silent")
}

func TestBuilderWait(t *testing.T) {
	cmd := isolate.New().
		Wait().
		BuildRun("./program")

	assertContains(t, cmd.Args, "--wait")
}

func TestBuilderMultipleVerbose(t *testing.T) {
	cmd := isolate.New().
		Verbose().
		Verbose().
		Verbose().
		BuildRun("./program")

	count := 0
	for _, arg := range cmd.Args {
		if arg == "--verbose" {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 --verbose flags, got %d", count)
	}
}

func TestBuilderIsolatePath(t *testing.T) {
	cmd := isolate.New().
		IsolatePath("/usr/local/bin/isolate").
		BuildInit()

	if cmd.Path != "/usr/local/bin/isolate" {
		t.Errorf("expected path /usr/local/bin/isolate, got %q", cmd.Path)
	}
}

func TestBuilderDefaultPath(t *testing.T) {
	cmd := isolate.New().BuildInit()

	if cmd.Path != "isolate" {
		t.Errorf("expected default path 'isolate', got %q", cmd.Path)
	}
}

func TestBuilderPrintCGRoot(t *testing.T) {
	cmd := isolate.New().
		BoxID(0).
		BuildPrintCGRoot()

	assertContains(t, cmd.Args, "--print-cg-root")
	assertContains(t, cmd.Args, "--box-id=0")
}

func TestCommandString(t *testing.T) {
	cmd := isolate.New().
		BoxID(0).
		MemoryLimit(256000).
		TimeLimit(5.0).
		BuildRun("./solution")

	s := cmd.String()
	if s == "" {
		t.Error("expected non-empty command string")
	}
	if !containsSubstring(s, "isolate") {
		t.Errorf("expected command string to contain 'isolate', got %q", s)
	}
	if !containsSubstring(s, "--run") {
		t.Errorf("expected command string to contain '--run', got %q", s)
	}
}

func TestBuilderCGInitAndCleanup(t *testing.T) {
	// CG flag should be included in init and cleanup
	initCmd := isolate.New().ControlGroup().BuildInit()
	assertContains(t, initCmd.Args, "--cg")

	cleanupCmd := isolate.New().ControlGroup().BuildCleanup()
	assertContains(t, cleanupCmd.Args, "--cg")
}

// Helper functions

func assertContains(t *testing.T, args []string, expected string) {
	t.Helper()
	for _, arg := range args {
		if arg == expected {
			return
		}
	}
	t.Errorf("expected args to contain %q, got %v", expected, args)
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringImpl(s, substr))
}

func containsSubstringImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
