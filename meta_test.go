package isolate_test

import (
	"strings"
	"testing"

	isolate "github.com/hoangnm/go-isolate"
)

func TestParseMetaSuccess(t *testing.T) {
	data := `time:1.234
time-wall:2.345
max-rss:12345
csw-forced:10
csw-voluntary:20
exitcode:0
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Time != 1.234 {
		t.Errorf("expected Time=1.234, got %g", meta.Time)
	}
	if meta.TimeWall != 2.345 {
		t.Errorf("expected TimeWall=2.345, got %g", meta.TimeWall)
	}
	if meta.MaxRSS != 12345 {
		t.Errorf("expected MaxRSS=12345, got %d", meta.MaxRSS)
	}
	if meta.CSWForced != 10 {
		t.Errorf("expected CSWForced=10, got %d", meta.CSWForced)
	}
	if meta.CSWVoluntary != 20 {
		t.Errorf("expected CSWVoluntary=20, got %d", meta.CSWVoluntary)
	}
	if meta.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got %d", meta.ExitCode)
	}
	if !meta.IsSuccess() {
		t.Error("expected IsSuccess()=true")
	}
}

func TestParseMetaRuntimeError(t *testing.T) {
	data := `status:RE
exitcode:1
message:Exited with error status 1
time:0.123
time-wall:0.234
max-rss:4096
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Status != isolate.StatusRuntimeError {
		t.Errorf("expected Status=RE, got %q", meta.Status)
	}
	if meta.ExitCode != 1 {
		t.Errorf("expected ExitCode=1, got %d", meta.ExitCode)
	}
	if meta.Message != "Exited with error status 1" {
		t.Errorf("expected specific message, got %q", meta.Message)
	}
	if meta.IsSuccess() {
		t.Error("expected IsSuccess()=false")
	}
}

func TestParseMetaSignal(t *testing.T) {
	data := `status:SG
exitsig:11
message:Caught fatal signal 11 (SIGSEGV)
time:0.050
time-wall:0.100
max-rss:2048
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Status != isolate.StatusSignal {
		t.Errorf("expected Status=SG, got %q", meta.Status)
	}
	if meta.ExitSignal != 11 {
		t.Errorf("expected ExitSignal=11, got %d", meta.ExitSignal)
	}
}

func TestParseMetaTimeout(t *testing.T) {
	data := `status:TO
killed:1
message:Time limit exceeded
time:5.000
time-wall:10.000
max-rss:65536
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Status != isolate.StatusTimeout {
		t.Errorf("expected Status=TO, got %q", meta.Status)
	}
	if !meta.Killed {
		t.Error("expected Killed=true")
	}
	if meta.IsSuccess() {
		t.Error("expected IsSuccess()=false for timeout")
	}
}

func TestParseMetaCG(t *testing.T) {
	data := `cg-mem:131072
cg-oom-killed:1
time:0.500
time-wall:1.000
exitcode:0
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.CGMem != 131072 {
		t.Errorf("expected CGMem=131072, got %d", meta.CGMem)
	}
	if !meta.CGOOMKilled {
		t.Error("expected CGOOMKilled=true")
	}
}

func TestParseMetaEmpty(t *testing.T) {
	meta, err := isolate.ParseMetaString("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !meta.IsSuccess() {
		t.Error("expected empty meta to be success")
	}
}

func TestParseMetaReader(t *testing.T) {
	r := strings.NewReader("exitcode:0\ntime:1.5\n")
	meta, err := isolate.ParseMeta(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got %d", meta.ExitCode)
	}
	if meta.Time != 1.5 {
		t.Errorf("expected Time=1.5, got %g", meta.Time)
	}
}

func TestParseMetaInvalidInt(t *testing.T) {
	_, err := isolate.ParseMetaString("exitcode:abc\n")
	if err == nil {
		t.Error("expected error for invalid int value")
	}
}

func TestParseMetaInvalidFloat(t *testing.T) {
	_, err := isolate.ParseMetaString("time:xyz\n")
	if err == nil {
		t.Error("expected error for invalid float value")
	}
}

func TestParseMetaInternalError(t *testing.T) {
	data := `status:XX
message:Internal error of the sandbox
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Status != isolate.StatusInternalError {
		t.Errorf("expected Status=XX, got %q", meta.Status)
	}
}

func TestParseMetaMessageWithColon(t *testing.T) {
	data := `message:Error: something went wrong
exitcode:1
`
	meta, err := isolate.ParseMetaString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Message != "Error: something went wrong" {
		t.Errorf("expected message with colon, got %q", meta.Message)
	}
}
