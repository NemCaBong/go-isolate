package isolate

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Status represents the two-letter status code from isolate's meta-file.
type Status string

const (
	// StatusOK indicates normal termination (no status in meta-file).
	StatusOK Status = ""
	// StatusRuntimeError indicates a non-zero exit code.
	StatusRuntimeError Status = "RE"
	// StatusSignal indicates the program died on a signal.
	StatusSignal Status = "SG"
	// StatusTimeout indicates a time limit exceeded.
	StatusTimeout Status = "TO"
	// StatusInternalError indicates an internal sandbox error.
	StatusInternalError Status = "XX"
)

// Meta represents parsed meta-file data produced by isolate.
type Meta struct {
	// CGMem is the total memory use by the control group (kilobytes).
	CGMem int
	// CGOOMKilled indicates the program was killed by the OOM killer.
	CGOOMKilled bool
	// CSWForced is the number of forced context switches.
	CSWForced int
	// CSWVoluntary is the number of voluntary context switches.
	CSWVoluntary int
	// ExitCode is the exit code if the program exited normally.
	ExitCode int
	// ExitSignal is the fatal signal number if the program was killed by a signal.
	ExitSignal int
	// Killed indicates the program was terminated by the sandbox.
	Killed bool
	// MaxRSS is the maximum resident set size (kilobytes).
	MaxRSS int
	// Message is a human-readable status message.
	Message string
	// Status is the two-letter status code (RE, SG, TO, XX).
	Status Status
	// Time is the CPU run time in seconds.
	Time float64
	// TimeWall is the wall clock time in seconds.
	TimeWall float64
}

// IsSuccess returns true if the program terminated normally with exit code 0.
func (m *Meta) IsSuccess() bool {
	return m.Status == StatusOK && m.ExitCode == 0 && !m.Killed
}

// ParseMeta parses an isolate meta-file from a reader.
func ParseMeta(r io.Reader) (*Meta, error) {
	meta := &Meta{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Split on first ':'
		idx := strings.IndexByte(line, ':')
		if idx < 0 {
			continue
		}

		key := line[:idx]
		value := line[idx+1:]

		var err error
		switch key {
		case "cg-mem":
			meta.CGMem, err = strconv.Atoi(value)
		case "cg-oom-killed":
			meta.CGOOMKilled = true
		case "csw-forced":
			meta.CSWForced, err = strconv.Atoi(value)
		case "csw-voluntary":
			meta.CSWVoluntary, err = strconv.Atoi(value)
		case "exitcode":
			meta.ExitCode, err = strconv.Atoi(value)
		case "exitsig":
			meta.ExitSignal, err = strconv.Atoi(value)
		case "killed":
			meta.Killed = true
		case "max-rss":
			meta.MaxRSS, err = strconv.Atoi(value)
		case "message":
			meta.Message = value
		case "status":
			meta.Status = Status(value)
		case "time":
			meta.Time, err = strconv.ParseFloat(value, 64)
		case "time-wall":
			meta.TimeWall, err = strconv.ParseFloat(value, 64)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to parse meta key %q value %q: %w", key, value, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read meta-file: %w", err)
	}

	return meta, nil
}

// ParseMetaString parses an isolate meta-file from a string.
func ParseMetaString(data string) (*Meta, error) {
	return ParseMeta(strings.NewReader(data))
}
