package isolate

import "path/filepath"

const (
	DefaultMetaFileName   = "metadata.txt"
	DefaultStdinFileName  = "stdin.txt"
	DefaultStdoutFileName = "stdout.txt"
	DefaultStderrFileName = "stderr.txt"
	DefaultIsolateName    = "isolate"
	SandBoxDirName        = "box"
)

// Builder provides a fluent API for constructing isolate commands.
// Use [New] to create a new Builder instance.
type Builder struct {
	// Path to the isolate binary. Defaults to "isolate".
	isolatePath string

	// Set after Init() is called.
	workDir    string
	sandboxDir string

	// --- Basic Options ---
	boxID  *int
	meta   string
	stdin  string
	stdout string
	stderr string

	stderrToStdout bool
	chdir          string
	verbose        int // number of -v flags
	silent         bool
	wait           bool

	// --- Limits ---
	memLimit      *int     // kilobytes
	timeLimit     *float64 // seconds
	wallTimeLimit *float64 // seconds
	extraTime     *float64 // seconds
	stackLimit    *int     // kilobytes
	openFiles     *int
	fsizeLimit    *int // kilobytes
	quotaBlocks   *int
	quotaInodes   *int
	coreLimit     *int // kilobytes
	processes     *int // nil = default (1), -1 = unlimited

	// --- Environment Rules ---
	fullEnv  bool
	envRules []EnvRule

	// --- Directory Rules ---
	noDefaultDirs bool
	dirRules      []DirRule

	// --- Control Groups ---
	cg    bool
	cgMem *int // kilobytes

	// --- Special Options ---
	shareNet     bool
	inheritFDs   bool
	ttyHack      bool
	specialFiles bool
	asUID        *int
	asGID        *int
}

// New creates a new [Builder] with default settings.
// The isolate binary path defaults to "isolate".
func New() *Builder {
	return &Builder{
		isolatePath: DefaultIsolateName,
		meta:        DefaultMetaFileName,
		stdin:       DefaultStdinFileName,
		stdout:      DefaultStdoutFileName,
		stderr:      DefaultStderrFileName,
	}
}

// WorkDir returns the isolate directory returned by --init (e.g., /var/local/lib/isolate/0).
// Empty before Init is called.
func (b *Builder) WorkDir() string { return b.workDir }

// SandboxDir returns the actual sandbox box directory (WorkDir/box).
// Empty before Init is called.
func (b *Builder) SandboxDir() string { return b.sandboxDir }

// setDirs is called by the Executor after --init completes to record the
// sandbox paths needed for building subsequent commands.
func (b *Builder) setDirs(isolateDir string) {
	b.workDir = isolateDir
	if isolateDir == "" {
		b.sandboxDir = ""
	} else {
		b.sandboxDir = filepath.Join(isolateDir, SandBoxDirName)
	}
}

// resolvePath converts a relative sandbox file path to an absolute host path
// using sandboxDir. Absolute paths and empty strings are returned unchanged.
func (b *Builder) resolvePath(rel string) string {
	if rel == "" || filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(b.sandboxDir, rel)
}

// IsolatePath sets the path to the isolate binary.
// Defaults to "isolate" (found via PATH).
func (b *Builder) IsolatePath(path string) *Builder {
	b.isolatePath = path
	return b
}

// --- Basic Options ---

// BoxID sets the sandbox ID (--box-id).
// Required when running multiple sandboxes in parallel.
// Defaults to 0 if not set.
func (b *Builder) BoxID(id int) *Builder {
	b.boxID = &id
	return b
}

// Meta sets the meta-file path for execution metadata (--meta).
func (b *Builder) Meta(file string) *Builder {
	b.meta = file
	return b
}

// Stdin redirects standard input from the given file (--stdin).
// The file must be present inside the sandbox (use WriteToSandbox to copy it).
func (b *Builder) Stdin(file string) *Builder {
	b.stdin = file
	return b
}

// Stdout redirects standard output to the given file (--stdout).
// The file will be created/written inside the sandbox.
func (b *Builder) Stdout(file string) *Builder {
	b.stdout = file
	return b
}

// Stderr redirects standard error to the given file (--stderr).
// The file will be created/written inside the sandbox. Mutually exclusive with StderrToStdout.
func (b *Builder) Stderr(file string) *Builder {
	b.stderr = file
	return b
}

// StderrToStdout redirects stderr to stdout (--stderr-to-stdout).
// Mutually exclusive with Stderr.
func (b *Builder) StderrToStdout() *Builder {
	b.stderrToStdout = true
	return b
}

// Chdir changes the working directory before execution (--chdir).
// The path must be relative to the sandbox root.
func (b *Builder) Chdir(dir string) *Builder {
	b.chdir = dir
	return b
}

// Verbose increases verbosity level (--verbose).
// Can be called multiple times to increase verbosity.
func (b *Builder) Verbose() *Builder {
	b.verbose++
	return b
}

// Silent sets the sandbox manager to keep silence (--silent).
func (b *Builder) Silent() *Builder {
	b.silent = true
	return b
}

// Wait makes the sandbox wait for other instances to finish (--wait).
func (b *Builder) Wait() *Builder {
	b.wait = true
	return b
}

// --- Limits ---

// MemoryLimit limits address space to size kilobytes (--mem).
func (b *Builder) MemoryLimit(sizeKB int) *Builder {
	b.memLimit = &sizeKB
	return b
}

// TimeLimit limits CPU run time to the given seconds (--time).
// Fractional values are allowed.
func (b *Builder) TimeLimit(seconds float64) *Builder {
	b.timeLimit = &seconds
	return b
}

// WallTimeLimit limits wall-clock time to the given seconds (--wall-time).
// Fractional values are allowed.
func (b *Builder) WallTimeLimit(seconds float64) *Builder {
	b.wallTimeLimit = &seconds
	return b
}

// ExtraTime sets the grace period after time limit is exceeded (--extra-time).
// Fractional values are allowed.
func (b *Builder) ExtraTime(seconds float64) *Builder {
	b.extraTime = &seconds
	return b
}

// StackLimit limits process stack to size kilobytes (--stack).
func (b *Builder) StackLimit(sizeKB int) *Builder {
	b.stackLimit = &sizeKB
	return b
}

// OpenFilesLimit limits number of open files (--open-files).
// Default is 64. Set to 0 for unlimited.
func (b *Builder) OpenFilesLimit(max int) *Builder {
	b.openFiles = &max
	return b
}

// FileSizeLimit limits size of each file created/modified (--fsize) in kilobytes.
func (b *Builder) FileSizeLimit(sizeKB int) *Builder {
	b.fsizeLimit = &sizeKB
	return b
}

// DiskQuota sets disk quota in blocks and inodes (--quota).
// Must be given to isolate --init.
func (b *Builder) DiskQuota(blocks, inodes int) *Builder {
	b.quotaBlocks = &blocks
	b.quotaInodes = &inodes
	return b
}

// CoreLimit limits core file size in kilobytes (--core).
// Defaults to 0 (no core files).
func (b *Builder) CoreLimit(sizeKB int) *Builder {
	b.coreLimit = &sizeKB
	return b
}

// Processes permits the program to create up to max processes/threads (--processes).
// Set to -1 for unlimited. Default is 1.
func (b *Builder) Processes(max int) *Builder {
	b.processes = &max
	return b
}

// ProcessesUnlimited permits an arbitrary number of processes/threads.
func (b *Builder) ProcessesUnlimited() *Builder {
	v := -1
	b.processes = &v
	return b
}

// --- Environment Rules ---

// FullEnv inherits all environment variables from the parent (--full-env).
func (b *Builder) FullEnv() *Builder {
	b.fullEnv = true
	return b
}

// InheritEnv inherits a specific variable from the parent (--env=var).
func (b *Builder) InheritEnv(variable string) *Builder {
	b.envRules = append(b.envRules, EnvRule{Variable: variable})
	return b
}

// SetEnv sets an environment variable to a specific value (--env=var=value).
func (b *Builder) SetEnv(variable, value string) *Builder {
	b.envRules = append(b.envRules, EnvRule{Variable: variable, Value: &value})
	return b
}

// RemoveEnv removes an environment variable (--env=var=).
func (b *Builder) RemoveEnv(variable string) *Builder {
	empty := ""
	b.envRules = append(b.envRules, EnvRule{Variable: variable, Value: &empty})
	return b
}

// --- Directory Rules ---

// NoDefaultDirs disables the default set of directory bindings (--no-default-dirs).
func (b *Builder) NoDefaultDirs() *Builder {
	b.noDefaultDirs = true
	return b
}

// Dir adds a directory binding rule (--dir).
// inside is the path as seen inside the sandbox,
// outside is the path as seen by the caller (empty to use /inside).
func (b *Builder) Dir(inside, outside string, options ...DirOption) *Builder {
	b.dirRules = append(b.dirRules, DirRule{
		Inside:  inside,
		Outside: outside,
		Options: options,
	})
	return b
}

// DirSimple binds a directory with the same inside and outside path (--dir=dir).
func (b *Builder) DirSimple(dir string, options ...DirOption) *Builder {
	b.dirRules = append(b.dirRules, DirRule{
		Inside:  dir,
		Options: options,
	})
	return b
}

// RemoveDir removes a directory rule (--dir=in=).
func (b *Builder) RemoveDir(inside string) *Builder {
	b.dirRules = append(b.dirRules, DirRule{
		Inside: inside,
		Remove: true,
	})
	return b
}

// --- Control Groups ---

// ControlGroup enables control group mode (--cg).
func (b *Builder) ControlGroup() *Builder {
	b.cg = true
	return b
}

// CGMemoryLimit limits total memory of the control group in kilobytes (--cg-mem).
func (b *Builder) CGMemoryLimit(sizeKB int) *Builder {
	b.cgMem = &sizeKB
	return b
}

// --- Special Options ---

// ShareNet keeps the child process in the parent's network namespace (--share-net).
func (b *Builder) ShareNet() *Builder {
	b.shareNet = true
	return b
}

// InheritFDs keeps all file descriptors from the parent (--inherit-fds).
func (b *Builder) InheritFDs() *Builder {
	b.inheritFDs = true
	return b
}

// TTYHack tries to handle interactive programs over a tty (--tty-hack).
func (b *Builder) TTYHack() *Builder {
	b.ttyHack = true
	return b
}

// SpecialFiles disables removal of special files created inside the sandbox (--special-files).
func (b *Builder) SpecialFiles() *Builder {
	b.specialFiles = true
	return b
}

// AsUID acts on behalf of the specified user ID (--as-uid).
// Only works if isolate was invoked by root.
func (b *Builder) AsUID(uid int) *Builder {
	b.asUID = &uid
	return b
}

// AsGID acts on behalf of the specified group ID (--as-gid).
// Only works if isolate was invoked by root.
func (b *Builder) AsGID(gid int) *Builder {
	b.asGID = &gid
	return b
}
