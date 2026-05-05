// Package isolate provides a Go library for building and executing
// isolate sandbox commands using the builder design pattern.
//
// Isolate is a tool for running processes inside a Linux sandbox (container)
// with strict resource limits and filesystem isolation. This library provides
// a type-safe, fluent API for constructing isolate commands.
//
// Basic usage:
//
//	exec := isolate.New().
//		BoxID(0).
//		MemoryLimit(256 * 1024).   // 256 MB
//		TimeLimit(5.0).            // 5 seconds
//		WallTimeLimit(10.0).       // 10 seconds
//		Stdin("input.txt").
//		Stdout("output.txt").
//		Stderr("error.txt").
//		Meta("/tmp/meta.txt").
//		Exec()
//
//	ctx := context.Background()
//
//	// 1. Initialize the sandbox
//	workDir, err := exec.Init(ctx)
//
//	// 2. IMPORTANT: Copy your program and data into the sandbox.
//	// The sandbox is an isolated environment and starts empty.
//	bin, _ := os.ReadFile("/path/to/solution")
//	err = exec.WriteToSandbox("solution", bin, 0755)
//
//	// 3. Run the program inside the sandbox
//	result, err := exec.Run(ctx, "./solution", "arg1", "arg2")
//
//	// 4. Clean up (no error return — cleanup is idempotent per isolate docs)
//	exec.Cleanup(ctx)
package isolate

import (
	"fmt"
	"strings"
)

// Action represents the isolate action to perform.
type Action int

const (
	// ActionInit initializes the sandbox.
	ActionInit Action = iota
	// ActionRun runs a program inside the sandbox.
	ActionRun
	// ActionCleanup cleans up the sandbox.
	ActionCleanup
)

// String returns the command-line flag for the action.
func (a Action) String() string {
	switch a {
	case ActionInit:
		return "--init"
	case ActionRun:
		return "--run"
	case ActionCleanup:
		return "--cleanup"
	default:
		return ""
	}
}

// DirOption represents options for a directory rule.
type DirOption string

const (
	// DirRW allows read-write access.
	DirRW DirOption = "rw"
	// DirDev allows access to character and block devices.
	DirDev DirOption = "dev"
	// DirNoExec disallows execution of binaries.
	DirNoExec DirOption = "noexec"
	// DirMaybe silently ignores the rule if the directory does not exist.
	DirMaybe DirOption = "maybe"
	// DirFS mounts a device-less filesystem instead of binding.
	DirFS DirOption = "fs"
	// DirTmp binds a freshly created temporary directory.
	DirTmp DirOption = "tmp"
	// DirNoRec does not bind recursively.
	DirNoRec DirOption = "norec"
)

// DirRule represents a single directory binding rule for the sandbox.
type DirRule struct {
	// Inside is the path as seen inside the sandbox.
	Inside string
	// Outside is the path as seen by the caller. Empty means same as Inside
	// (with "/" prepended). Set to empty string with remove=true to remove a rule.
	Outside string
	// Options are the directory options (rw, dev, noexec, maybe, fs, tmp, norec).
	Options []DirOption
	// Remove indicates this rule removes a previously set directory rule.
	Remove bool
}

// String formats the directory rule as a command-line argument value.
func (d DirRule) String() string {
	if d.Remove {
		return fmt.Sprintf("%s=", d.Inside)
	}

	var sb strings.Builder
	if d.Outside != "" && d.Outside != d.Inside {
		sb.WriteString(fmt.Sprintf("%s=%s", d.Inside, d.Outside))
	} else {
		sb.WriteString(d.Inside)
	}

	if len(d.Options) > 0 {
		opts := make([]string, len(d.Options))
		for i, o := range d.Options {
			opts[i] = string(o)
		}
		sb.WriteString(":")
		sb.WriteString(strings.Join(opts, ","))
	}

	return sb.String()
}

// EnvRule represents an environment variable rule.
type EnvRule struct {
	// Variable is the environment variable name.
	Variable string
	// Value is the value to set. If nil, the variable is inherited from parent.
	Value *string
}

// String formats the environment rule as a command-line argument value.
func (e EnvRule) String() string {
	if e.Value != nil {
		return fmt.Sprintf("%s=%s", e.Variable, *e.Value)
	}
	return e.Variable
}
