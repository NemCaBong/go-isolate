package isolate

import (
	"fmt"
	"strings"
)

// Command represents a fully built isolate command ready for execution.
type Command struct {
	// Path is the path to the isolate binary.
	Path string
	// Args is the list of command-line arguments (excluding the binary path).
	Args []string
}

// String returns the full command line as a string.
func (c *Command) String() string {
	parts := append([]string{c.Path}, c.Args...)
	return strings.Join(parts, " ")
}

// buildCommonArgs builds command-line arguments that are common across actions.
func (b *Builder) buildCommonArgs() []string {
	var args []string

	// Box ID
	if b.boxID != nil {
		args = append(args, fmt.Sprintf("--box-id=%d", *b.boxID))
	}

	// Verbose
	for i := 0; i < b.verbose; i++ {
		args = append(args, "--verbose")
	}

	// Silent
	if b.silent {
		args = append(args, "--silent")
	}

	// Wait
	if b.wait {
		args = append(args, "--wait")
	}

	// Control Groups
	if b.cg {
		args = append(args, "--cg")
	}

	return args
}

// buildRunArgs builds command-line arguments specific to the --run action.
func (b *Builder) buildRunArgs() []string {
	var args []string

	// I/O redirection — resolve relative paths to absolute host paths via sandboxDir.
	if b.stdin != "" {
		args = append(args, fmt.Sprintf("--stdin=%s", b.resolvePath(b.stdin)))
	}
	if b.stdout != "" {
		args = append(args, fmt.Sprintf("--stdout=%s", b.resolvePath(b.stdout)))
	}
	if b.stderr != "" {
		args = append(args, fmt.Sprintf("--stderr=%s", b.resolvePath(b.stderr)))
	}
	if b.stderrToStdout {
		args = append(args, "--stderr-to-stdout")
	}

	// Meta is run-only (captures timing/resource info) and needs a host-side absolute path.
	if b.meta != "" {
		args = append(args, fmt.Sprintf("--meta=%s", b.resolvePath(b.meta)))
	}

	// Chdir
	if b.chdir != "" {
		args = append(args, fmt.Sprintf("--chdir=%s", b.chdir))
	}

	// Limits
	if b.memLimit != nil {
		args = append(args, fmt.Sprintf("--mem=%d", *b.memLimit))
	}
	if b.timeLimit != nil {
		args = append(args, fmt.Sprintf("--time=%g", *b.timeLimit))
	}
	if b.wallTimeLimit != nil {
		args = append(args, fmt.Sprintf("--wall-time=%g", *b.wallTimeLimit))
	}
	if b.extraTime != nil {
		args = append(args, fmt.Sprintf("--extra-time=%g", *b.extraTime))
	}
	if b.stackLimit != nil {
		args = append(args, fmt.Sprintf("--stack=%d", *b.stackLimit))
	}
	if b.openFiles != nil {
		args = append(args, fmt.Sprintf("--open-files=%d", *b.openFiles))
	}
	if b.fsizeLimit != nil {
		args = append(args, fmt.Sprintf("--fsize=%d", *b.fsizeLimit))
	}
	if b.coreLimit != nil {
		args = append(args, fmt.Sprintf("--core=%d", *b.coreLimit))
	}
	if b.processes != nil {
		if *b.processes == -1 {
			args = append(args, "--processes")
		} else {
			args = append(args, fmt.Sprintf("--processes=%d", *b.processes))
		}
	}

	// CG memory limit
	if b.cgMem != nil {
		args = append(args, fmt.Sprintf("--cg-mem=%d", *b.cgMem))
	}

	// Environment rules
	if b.fullEnv {
		args = append(args, "--full-env")
	}
	for _, rule := range b.envRules {
		args = append(args, fmt.Sprintf("--env=%s", rule.String()))
	}

	// Directory rules
	if b.noDefaultDirs {
		args = append(args, "--no-default-dirs")
	}
	for _, rule := range b.dirRules {
		args = append(args, fmt.Sprintf("--dir=%s", rule.String()))
	}

	// Special options
	if b.shareNet {
		args = append(args, "--share-net")
	}
	if b.inheritFDs {
		args = append(args, "--inherit-fds")
	}
	if b.ttyHack {
		args = append(args, "--tty-hack")
	}
	if b.specialFiles {
		args = append(args, "--special-files")
	}
	if b.asUID != nil {
		args = append(args, fmt.Sprintf("--as-uid=%d", *b.asUID))
	}
	if b.asGID != nil {
		args = append(args, fmt.Sprintf("--as-gid=%d", *b.asGID))
	}

	return args
}

// buildInitArgs builds command-line arguments specific to the --init action.
func (b *Builder) buildInitArgs() []string {
	var args []string

	// Quota is only valid with --init
	if b.quotaBlocks != nil && b.quotaInodes != nil {
		args = append(args, fmt.Sprintf("--quota=%d,%d", *b.quotaBlocks, *b.quotaInodes))
	}

	return args
}

// BuildInit builds the command for sandbox initialization (--init).
func (b *Builder) BuildInit() *Command {
	var args []string
	args = append(args, b.buildCommonArgs()...)
	args = append(args, b.buildInitArgs()...)
	args = append(args, ActionInit.String())

	return &Command{
		Path: b.isolatePath,
		Args: args,
	}
}

// BuildRun builds the command for running a program (--run).
// program is the executable to run, and programArgs are its arguments.
// Call Init first so that sandboxDir is set and I/O paths resolve correctly.
func (b *Builder) BuildRun(program string, programArgs ...string) *Command {
	var args []string
	args = append(args, b.buildCommonArgs()...)
	args = append(args, b.buildRunArgs()...)
	args = append(args, ActionRun.String())
	args = append(args, "--")
	args = append(args, program)
	args = append(args, programArgs...)

	return &Command{
		Path: b.isolatePath,
		Args: args,
	}
}

// BuildCleanup builds the command for sandbox cleanup (--cleanup).
func (b *Builder) BuildCleanup() *Command {
	var args []string
	args = append(args, b.buildCommonArgs()...)
	args = append(args, ActionCleanup.String())

	return &Command{
		Path: b.isolatePath,
		Args: args,
	}
}

// BuildPrintCGRoot builds the command to print the control group root (--print-cg-root).
func (b *Builder) BuildPrintCGRoot() *Command {
	var args []string
	if b.boxID != nil {
		args = append(args, fmt.Sprintf("--box-id=%d", *b.boxID))
	}
	args = append(args, "--print-cg-root")

	return &Command{
		Path: b.isolatePath,
		Args: args,
	}
}
