# Development Rules
- **Single Responsibility Principle (SRP):** Every file must have a single, well-defined purpose.
- **Test Organization:** Tests MUST use Mocks for all external interfaces to remain fast, deterministic, and side-effect free.

# Goal Description
Update `gorun.KillAllByName` to be fully cross-platform compatible, particularly ensuring that the Windows implementation reliably targets executables by appending the `.exe` extension if it is missing. This ensures daemons can be reliably shut down by name on Windows without relying on exact user input.

# Proposed Changes

## Component: gorun

### [MODIFY] cleanup.go
- Inside `killAllWindows(executableName string) error`:
  - Check if `executableName` ends with `.exe` (case-insensitive).
  - If it does not end with `.exe`, append `.exe` to `executableName` before passing it to the `taskkill` command.
  - This ensures `taskkill /IM <name>.exe` works robustly on Windows even when the caller provides just the command name (e.g., `tinywasm` instead of `tinywasm.exe`).

### [MODIFY] cleanup_test.go
- Add a test case for `killAllWindows` (if feasible to test independently or mock) to verify that the `.exe` extension is appended correctly. Because `killAllWindows` uses `exec.Command`, we might just verify the logic or trust the standard implementation if mocking `exec` is not over-engineered. The focus should be on unit-testing the string manipulation if extracted into a helper, or just updating the existing tests.

## Verification Plan

### Automated Tests
1. Run `gotest` in the `gorun` root to ensure all existing tests pass and no regressions are introduced.

### Manual Verification
1. On a Windows environment (or a mocked Windows test), verify that calling `KillAllByName("mydaemon", false)` successfully executes `taskkill /F /IM mydaemon.exe`.
