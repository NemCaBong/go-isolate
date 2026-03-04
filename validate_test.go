package isolate_test

import (
	"testing"

	isolate "github.com/NemCaBong/go-isolate"
)

func TestValidateSuccess(t *testing.T) {
	b := isolate.New().
		BoxID(0).
		MemoryLimit(256000).
		TimeLimit(5.0).
		WallTimeLimit(10.0).
		ExtraTime(0.5).
		StackLimit(8192).
		OpenFilesLimit(128).
		FileSizeLimit(1024).
		CoreLimit(0).
		Processes(10)

	if err := b.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateStderrConflict(t *testing.T) {
	b := isolate.New().
		Stderr("error.txt").
		StderrToStdout()

	if err := b.Validate(); err == nil {
		t.Error("expected error for Stderr + StderrToStdout")
	}
}

func TestValidateNegativeBoxID(t *testing.T) {
	b := isolate.New().BoxID(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative box ID")
	}
}

func TestValidateNegativeMemoryLimit(t *testing.T) {
	b := isolate.New().MemoryLimit(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative memory limit")
	}
}

func TestValidateZeroTimeLimit(t *testing.T) {
	b := isolate.New().TimeLimit(0)

	if err := b.Validate(); err == nil {
		t.Error("expected error for zero time limit")
	}
}

func TestValidateNegativeWallTime(t *testing.T) {
	b := isolate.New().WallTimeLimit(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative wall time")
	}
}

func TestValidateNegativeExtraTime(t *testing.T) {
	b := isolate.New().ExtraTime(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative extra time")
	}
}

func TestValidateNegativeStackLimit(t *testing.T) {
	b := isolate.New().StackLimit(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative stack limit")
	}
}

func TestValidateNegativeOpenFiles(t *testing.T) {
	b := isolate.New().OpenFilesLimit(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative open files limit")
	}
}

func TestValidateNegativeFileSize(t *testing.T) {
	b := isolate.New().FileSizeLimit(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative file size limit")
	}
}

func TestValidatePartialQuota(t *testing.T) {
	// The DiskQuota method sets both at once, so we'd need to test
	// at initialization level. Since DiskQuota always sets both,
	// this case cannot actually occur through the builder API.
	// We'll test that valid quota passes.
	b := isolate.New().DiskQuota(1000, 500)

	if err := b.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateNegativeCoreLimit(t *testing.T) {
	b := isolate.New().CoreLimit(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative core limit")
	}
}

func TestValidateZeroCoreLimit(t *testing.T) {
	b := isolate.New().CoreLimit(0)

	if err := b.Validate(); err != nil {
		t.Errorf("expected no error for zero core limit, got %v", err)
	}
}

func TestValidateZeroProcesses(t *testing.T) {
	b := isolate.New().Processes(0)

	if err := b.Validate(); err == nil {
		t.Error("expected error for zero processes")
	}
}

func TestValidateProcessesUnlimited(t *testing.T) {
	b := isolate.New().ProcessesUnlimited()

	if err := b.Validate(); err != nil {
		t.Errorf("expected no error for unlimited processes, got %v", err)
	}
}

func TestValidateCGMemWithoutCG(t *testing.T) {
	b := isolate.New().CGMemoryLimit(256000)

	if err := b.Validate(); err == nil {
		t.Error("expected error for CG memory limit without CG mode")
	}
}

func TestValidateCGMemWithCG(t *testing.T) {
	b := isolate.New().ControlGroup().CGMemoryLimit(256000)

	if err := b.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateNegativeCGMem(t *testing.T) {
	b := isolate.New().ControlGroup().CGMemoryLimit(-1)

	if err := b.Validate(); err == nil {
		t.Error("expected error for negative CG memory limit")
	}
}

func TestValidateMinimalBuilder(t *testing.T) {
	b := isolate.New()

	if err := b.Validate(); err != nil {
		t.Errorf("expected no error for minimal builder, got %v", err)
	}
}

func TestValidateZeroOpenFiles(t *testing.T) {
	// 0 means unlimited open files
	b := isolate.New().OpenFilesLimit(0)

	if err := b.Validate(); err != nil {
		t.Errorf("expected no error for zero open files (unlimited), got %v", err)
	}
}
