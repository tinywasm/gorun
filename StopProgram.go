package gorun

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"
)

func (h *GoRun) StopProgram() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.KillAllOnStop {
		return h.stopProgramAndCleanupUnsafe(true)
	}
	return h.stopProgramUnsafe()
}

// stopProgramUnsafe stops the program without acquiring the mutex
// Should only be called when mutex is already held
func (h *GoRun) stopProgramUnsafe() error {
	if !h.isRunning || h.Cmd == nil || h.Cmd.Process == nil {
		h.isRunning = false
		return nil
	}

	// Check if process has already exited
	if h.Cmd.ProcessState != nil && h.Cmd.ProcessState.Exited() {
		h.isRunning = false
		return nil
	}

	// If Wait() has already been called, don't try to wait again
	if h.hasWaited {
		h.isRunning = false
		return nil
	}

	h.isRunning = false
	err := killProcessGraceful(h.Cmd.Process)
	h.hasWaited = true
	return err
}

// killProcessGraceful terminates a single process.
// Windows: immediate Kill(). Unix: SIGTERM -> 3s graceful wait -> SIGKILL.
func killProcessGraceful(process *os.Process) error {
	// Cross-platform graceful shutdown approach
	if runtime.GOOS == "windows" {
		// On Windows, we don't have SIGTERM, so we use Kill directly
		if err := process.Kill(); err != nil {
			if err.Error() == "os: process already finished" {
				return nil
			}
			return err
		}
		return nil
	}

	// On Unix-like systems (Linux, macOS), try graceful shutdown first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// If SIGTERM fails, it could be because the process already exited
		// Check if it's an "os: process already finished" error
		if err.Error() == "os: process already finished" {
			return nil
		}
		// For other errors, try force kill
		if killErr := process.Kill(); killErr != nil {
			if killErr.Error() == "os: process already finished" {
				return nil
			}
			return killErr
		}
		return nil
	}

	// Wait a bit for graceful shutdown
	done := make(chan error, 1)
	go func() {
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-time.After(3 * time.Second):
		// Timeout reached, force kill
		fmt.Fprintf(os.Stderr, "Process did not terminate gracefully, forcing kill\n")
		if err := process.Kill(); err != nil {
			// If kill fails with "process already finished", that's not an error
			if err.Error() == "os: process already finished" {
				return nil
			}
			return err
		}
		// Wait for the process to actually die after kill
		<-done
		return nil
	case err := <-done:
		// Process terminated gracefully
		// If we get "no child processes" error, it means the process already exited
		if err != nil && (err.Error() == "waitid: no child processes" || err.Error() == "wait: no child processes") {
			return nil // This is not an error, the process already exited
		}
		return err
	}
}
