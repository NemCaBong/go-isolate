package isolate

import "fmt"

// Validate checks the builder configuration for common errors.
// It returns an error if any configuration is invalid.
func (b *Builder) Validate() error {
	// Cannot use both --stderr and --stderr-to-stdout
	if b.stderr != "" && b.stderrToStdout {
		return fmt.Errorf("cannot use both Stderr and StderrToStdout: they are mutually exclusive")
	}

	// Box ID should be non-negative
	if b.boxID != nil && *b.boxID < 0 {
		return fmt.Errorf("box ID must be non-negative, got %d", *b.boxID)
	}

	// Memory limit should be positive
	if b.memLimit != nil && *b.memLimit <= 0 {
		return fmt.Errorf("memory limit must be positive, got %d", *b.memLimit)
	}

	// Time limit should be positive
	if b.timeLimit != nil && *b.timeLimit <= 0 {
		return fmt.Errorf("time limit must be positive, got %g", *b.timeLimit)
	}

	// Wall time limit should be positive
	if b.wallTimeLimit != nil && *b.wallTimeLimit <= 0 {
		return fmt.Errorf("wall time limit must be positive, got %g", *b.wallTimeLimit)
	}

	// Extra time should be non-negative
	if b.extraTime != nil && *b.extraTime < 0 {
		return fmt.Errorf("extra time must be non-negative, got %g", *b.extraTime)
	}

	// Stack limit should be positive
	if b.stackLimit != nil && *b.stackLimit <= 0 {
		return fmt.Errorf("stack limit must be positive, got %d", *b.stackLimit)
	}

	// Open files should be non-negative
	if b.openFiles != nil && *b.openFiles < 0 {
		return fmt.Errorf("open files limit must be non-negative, got %d", *b.openFiles)
	}

	// File size limit should be positive
	if b.fsizeLimit != nil && *b.fsizeLimit <= 0 {
		return fmt.Errorf("file size limit must be positive, got %d", *b.fsizeLimit)
	}

	// Quota: both or neither must be set
	if (b.quotaBlocks != nil) != (b.quotaInodes != nil) {
		return fmt.Errorf("both quota blocks and inodes must be set together")
	}

	// Quota values should be positive
	if b.quotaBlocks != nil && *b.quotaBlocks <= 0 {
		return fmt.Errorf("quota blocks must be positive, got %d", *b.quotaBlocks)
	}
	if b.quotaInodes != nil && *b.quotaInodes <= 0 {
		return fmt.Errorf("quota inodes must be positive, got %d", *b.quotaInodes)
	}

	// Core limit should be non-negative
	if b.coreLimit != nil && *b.coreLimit < 0 {
		return fmt.Errorf("core limit must be non-negative, got %d", *b.coreLimit)
	}

	// Processes should be -1 (unlimited) or positive
	if b.processes != nil && *b.processes < -1 {
		return fmt.Errorf("processes must be -1 (unlimited) or positive, got %d", *b.processes)
	}
	if b.processes != nil && *b.processes == 0 {
		return fmt.Errorf("processes must be -1 (unlimited) or positive, got 0")
	}

	// CG memory limit requires CG mode
	if b.cgMem != nil && !b.cg {
		return fmt.Errorf("CG memory limit requires control group mode to be enabled")
	}

	// CG memory limit should be positive
	if b.cgMem != nil && *b.cgMem <= 0 {
		return fmt.Errorf("CG memory limit must be positive, got %d", *b.cgMem)
	}

	return nil
}
