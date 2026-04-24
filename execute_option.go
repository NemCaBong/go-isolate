package isolate

type ExecuteOption func(*Executor)

// --- Basic Options ---
func WithStdin(file string) ExecuteOption {
	return func(e *Executor) {
		e.builder.Stdin(file)
	}
}

func WithStdout(file string) ExecuteOption {
	return func(e *Executor) {
		e.builder.Stdout(file)
	}
}

func WithStderr(file string) ExecuteOption {
	return func(e *Executor) {
		e.builder.Stderr(file)
	}
}

func WithStderrToStdout() ExecuteOption {
	return func(e *Executor) {
		e.builder.StderrToStdout()
	}
}

func WithChdir(dir string) ExecuteOption {
	return func(e *Executor) {
		e.builder.Chdir(dir)
	}
}

func WithVerbose() ExecuteOption {
	return func(e *Executor) {
		e.builder.Verbose()
	}
}

func WithSilent() ExecuteOption {
	return func(e *Executor) {
		e.builder.Silent()
	}
}

func WithWait() ExecuteOption {
	return func(e *Executor) {
		e.builder.Wait()
	}
}

// --- Limits ---

func WithMemoryLimit(sizeKB int) ExecuteOption {
	return func(e *Executor) {
		e.builder.MemoryLimit(sizeKB)
	}
}

func WithTimeLimit(seconds float64) ExecuteOption {
	return func(e *Executor) {
		e.builder.TimeLimit(seconds)
	}
}

func WithWallTimeLimit(seconds float64) ExecuteOption {
	return func(e *Executor) {
		e.builder.WallTimeLimit(seconds)
	}
}

func WithExtraTime(seconds float64) ExecuteOption {
	return func(e *Executor) {
		e.builder.ExtraTime(seconds)
	}
}

func WithStackLimit(sizeKB int) ExecuteOption {
	return func(e *Executor) {
		e.builder.StackLimit(sizeKB)
	}
}

func WithOpenFilesLimit(max int) ExecuteOption {
	return func(e *Executor) {
		e.builder.OpenFilesLimit(max)
	}
}

func WithFileSizeLimit(sizeKB int) ExecuteOption {
	return func(e *Executor) {
		e.builder.FileSizeLimit(sizeKB)
	}
}

func WithDiskQuota(blocks, inodes int) ExecuteOption {
	return func(e *Executor) {
		e.builder.DiskQuota(blocks, inodes)
	}
}

func WithCoreLimit(sizeKB int) ExecuteOption {
	return func(e *Executor) {
		e.builder.CoreLimit(sizeKB)
	}
}

func WithProcesses(max int) ExecuteOption {
	return func(e *Executor) {
		e.builder.Processes(max)
	}
}

func WithProcessesUnlimited() ExecuteOption {
	return func(e *Executor) {
		e.builder.ProcessesUnlimited()
	}
}

// --- Environment Rules ---

func WithFullEnv() ExecuteOption {
	return func(e *Executor) {
		e.builder.FullEnv()
	}
}

func WithInheritEnv(variable string) ExecuteOption {
	return func(e *Executor) {
		e.builder.InheritEnv(variable)
	}
}

func WithSetEnv(variable, value string) ExecuteOption {
	return func(e *Executor) {
		e.builder.SetEnv(variable, value)
	}
}

func WithRemoveEnv(variable string) ExecuteOption {
	return func(e *Executor) {
		e.builder.RemoveEnv(variable)
	}
}

// --- Directory Rules ---

func WithNoDefaultDirs() ExecuteOption {
	return func(e *Executor) {
		e.builder.NoDefaultDirs()
	}
}

func WithDir(inside, outside string, options ...DirOption) ExecuteOption {
	return func(e *Executor) {
		e.builder.Dir(inside, outside, options...)
	}
}

func WithDirSimple(dir string, options ...DirOption) ExecuteOption {
	return func(e *Executor) {
		e.builder.DirSimple(dir, options...)
	}
}

func WithRemoveDir(inside string) ExecuteOption {
	return func(e *Executor) {
		e.builder.RemoveDir(inside)
	}
}

// --- Control Groups ---

func WithControlGroup() ExecuteOption {
	return func(e *Executor) {
		e.builder.ControlGroup()
	}
}

func WithCGMemoryLimit(sizeKB int) ExecuteOption {
	return func(e *Executor) {
		e.builder.CGMemoryLimit(sizeKB)
	}
}

// --- Special Options ---

func WithShareNet() ExecuteOption {
	return func(e *Executor) {
		e.builder.ShareNet()
	}
}

func WithInheritFDs() ExecuteOption {
	return func(e *Executor) {
		e.builder.InheritFDs()
	}
}

func WithTTYHack() ExecuteOption {
	return func(e *Executor) {
		e.builder.TTYHack()
	}
}

func WithSpecialFiles() ExecuteOption {
	return func(e *Executor) {
		e.builder.SpecialFiles()
	}
}

func WithAsUID(uid int) ExecuteOption {
	return func(e *Executor) {
		e.builder.AsUID(uid)
	}
}

func WithAsGID(gid int) ExecuteOption {
	return func(e *Executor) {
		e.builder.AsGID(gid)
	}
}
