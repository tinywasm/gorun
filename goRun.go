package gorun

import (
	"os/exec"
	"sync"
)

type Config struct {
	ExecProgramPath      string          // eg: "server/main.exe"
	RunArguments         func() []string // eg: []string{"dev"}
	ExitChan             chan bool
	Logger               func(message ...any)
	KillAllOnStop        bool   // If true, kills all instances of the executable when stopping
	DisableGlobalCleanup bool   // If true, disables global cleanup (pgrep -f) even if KillAllOnStop is true
	WorkingDir           string // eg: "/path/to/working/dir"
}

type GoRun struct {
	*Config
	Cmd        *exec.Cmd
	isRunning  bool
	hasWaited  bool         // Track if Wait() has been called
	mutex      sync.RWMutex // Protect concurrent access to running state
	safeBuffer *SafeBuffer  // Thread-safe buffer for Logger
}

func New(c *Config) *GoRun {
	var buffer *SafeBuffer
	if c.Logger != nil {
		// Create SafeBuffer that forwards to the function logger
		buffer = NewSafeBufferWithForward(c.Logger)
	} else {
		buffer = NewSafeBuffer()
	}

	return &GoRun{
		Config:     c,
		Cmd:        &exec.Cmd{},
		isRunning:  false,
		hasWaited:  false,
		mutex:      sync.RWMutex{},
		safeBuffer: buffer,
	}
}

// getOutput returns the captured output in a thread-safe manner (unexported)
func (h *GoRun) getOutput() string {
	return h.safeBuffer.String()
}
