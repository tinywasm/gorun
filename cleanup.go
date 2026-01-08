package gorun

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// KillAllByName kills all running processes that match the given executable name
// This is useful for cleanup when multiple instances might be running
func KillAllByName(executableName string, disableGlobal bool) error {
	// SAFETY: During tests, we might want to disable global cleanup to avoid
	// accidentally killing the IDE or other unrelated processes.
	if disableGlobal {
		return nil
	}

	switch runtime.GOOS {
	case "windows":
		return killAllWindows(executableName)
	default:
		return killAllUnix(executableName)
	}
}

// killAllWindows kills all processes by name on Windows
func killAllWindows(executableName string) error {
	// Use taskkill command on Windows
	cmd := exec.Command("taskkill", "/F", "/IM", executableName)
	output, err := cmd.CombinedOutput()

	// taskkill returns error if no processes found, which is ok
	if err != nil && !strings.Contains(string(output), "not found") {
		return fmt.Errorf("failed to kill processes %s: %v, output: %s", executableName, err, output)
	}

	return nil
}

// killAllUnix kills all processes by name on Unix-like systems (Linux, macOS)
func killAllUnix(executableName string) error {
	// First, find all PIDs matching the executable name
	cmd := exec.Command("pgrep", "-f", executableName)
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns 1 if no processes found, which is ok
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return nil // No processes found
		}
		return fmt.Errorf("failed to find processes %s: %v", executableName, err)
	}

	// Parse PIDs and kill them
	pids := strings.Fields(string(output))
	var errors []string

	myPid := os.Getpid()
	parentPid := os.Getppid()

	for _, pidStr := range pids {
		pid, err := strconv.Atoi(strings.TrimSpace(pidStr))
		if err != nil {
			errors = append(errors, fmt.Sprintf("invalid PID %s: %v", pidStr, err))
			continue
		}

		// SAFETY: Never kill ourselves or our parent (IDE)
		if pid == myPid || pid == parentPid {
			continue
		}

		// Find the process and kill it
		process, err := os.FindProcess(pid)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to find process %d: %v", pid, err))
			continue
		}

		// Try graceful kill first (SIGTERM), then force kill if needed
		if err := process.Signal(os.Interrupt); err != nil {
			if err := process.Kill(); err != nil {
				errors = append(errors, fmt.Sprintf("failed to kill process %d: %v", pid, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some processes could not be killed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// StopProgramAndCleanup stops the current program and optionally kills all instances
// of the same executable name
func (h *GoRun) StopProgramAndCleanup(killAll bool) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.stopProgramAndCleanupUnsafe(killAll)
}

// stopProgramAndCleanupUnsafe is the unsafe version that doesn't acquire mutex
func (h *GoRun) stopProgramAndCleanupUnsafe(killAll bool) error {
	// First stop our specific process
	err := h.stopProgramUnsafe()

	// If requested, also kill all other instances (safety check inside KillAllByName)
	if killAll && h.ExecProgramPath != "" {
		// Extract executable name from path
		execName := h.ExecProgramPath
		if lastSlash := strings.LastIndex(execName, "/"); lastSlash != -1 {
			execName = execName[lastSlash+1:]
		}
		if lastBackslash := strings.LastIndex(execName, "\\"); lastBackslash != -1 {
			execName = execName[lastBackslash+1:]
		}

		if cleanupErr := KillAllByName(execName, h.DisableGlobalCleanup); cleanupErr != nil {
			// Log the cleanup error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup all instances of %s: %v\n", execName, cleanupErr)
		}
	}

	return err
}
