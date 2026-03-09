# Development Rules
- **Single Responsibility Principle (SRP):** Every file must have a single, well-defined purpose.
- **Standard Library Only:** NEVER use external assertion libraries. Use only the standard `testing` package.

# Goal Description
Expose a clean public API `StopApp(name string) error` that terminates all running instances of a daemon by name. Eliminates duplicated kill logic by extracting the cross-platform process kill from `stopProgramUnsafe` into free functions, then reusing them in both `StopProgram` and `killAllUnix`.

## Problem
1. `KillAllByName` leaks the `disableGlobal` testing concern into the public API.
2. On Windows, `taskkill /IM` requires the `.exe` extension, but callers pass just the name.
3. `killAllUnix` duplicates kill logic (primitive `os.Interrupt` → `process.Kill()`) instead of reusing `StopProgram`'s robust implementation (SIGTERM → 3s timeout → SIGKILL).

## Proposed Changes

### [MODIFY] [StopProgram.go](StopProgram.go)

#### Extract `killProcessGraceful` — free function (unexported)
Extract the cross-platform kill logic from `stopProgramUnsafe` (lines 44-105) into a free function:
```go
// killProcessGraceful terminates a single process.
// Windows: immediate Kill(). Unix: SIGTERM → 3s graceful wait → SIGKILL.
func killProcessGraceful(process *os.Process) error {
    // (move existing logic from stopProgramUnsafe lines 44-105 here)
}
```

#### Simplify `stopProgramUnsafe` — delegate to `killProcessGraceful`
```go
func (h *GoRun) stopProgramUnsafe() error {
    if !h.isRunning || h.Cmd == nil || h.Cmd.Process == nil {
        h.isRunning = false
        return nil
    }
    if h.Cmd.ProcessState != nil && h.Cmd.ProcessState.Exited() {
        h.isRunning = false
        return nil
    }
    if h.hasWaited {
        h.isRunning = false
        return nil
    }

    h.isRunning = false
    h.hasWaited = true
    return killProcessGraceful(h.Cmd.Process)
}
```
The public API (`StopProgram`, `StopProgramAndCleanup`) remains unchanged.

---

### [MODIFY] [cleanup.go](cleanup.go)

#### 1. Add `StopApp` — delegates to existing functions
```go
// StopApp terminates all running instances of the named daemon/service.
// Detects the OS and delegates to the existing platform-specific kill functions.
func StopApp(name string) error {
    switch runtime.GOOS {
    case "windows":
        return killAllWindows(name)
    case "linux", "darwin":
        return killAllUnix(name)
    default:
        return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
    }
}
```

#### 2. Fix `killAllWindows` — auto-append `.exe`
Add at the top of the existing function:
```go
if !strings.HasSuffix(strings.ToLower(executableName), ".exe") {
    executableName += ".exe"
}
```

#### 3. Refactor `killAllUnix` — reuse `killProcessGraceful`
Replace the primitive kill block (lines 76-88) with a call to the extracted function:
```go
// Before (duplicated logic):
if err := process.Signal(os.Interrupt); err != nil {
    if err := process.Kill(); err != nil { ... }
}

// After (reuses extracted function):
if err := killProcessGraceful(process); err != nil {
    errors = append(errors, fmt.Sprintf("failed to kill process %d: %v", pid, err))
}
```

### [MODIFY] [cleanup_test.go](cleanup_test.go)

Add:
```go
func TestStopApp_NonExistent(t *testing.T) {
    err := StopApp("nonexistent_program_12345")
    if err != nil {
        t.Errorf("StopApp should not error for non-existent program: %v", err)
    }
}
```

## Verification Plan

### Automated Tests
```bash
cd gorun && go test -v -race ./...
```
- All existing `StopProgram` tests pass (no regressions — same logic, just moved to free function).
- All existing cleanup tests pass.
- `TestStopApp_NonExistent` passes.

