> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Part of the orchestrator: `../../docs/HOT_RELOAD_MASTER_PLAN.md` (Phase B). Runs in parallel with `gobuild/docs/PLAN.md` and `depfind/docs/PLAN.md`; no dependencies.

# gorun — extract a `Runner` interface for fast unit testing

## Problem

`server`'s `externalStrategy` depends directly on the concrete struct
`*gorun.GoRun` to run/stop the compiled server binary. All existing `gorun`
tests exercise real compiled subprocesses from `gorun/testdata/*.go` — valid
for `gorun` itself, but it means `server` has no way to unit-test its
restart/reload orchestration logic (e.g. "on file event X, was RunProgram
called after CompileProgram, exactly once") without spawning real processes.

## Required change

1. In `gorun/goRun.go` (or a new `gorun/runner.go`), define a minimal
   interface capturing the public surface `server/strategies.go` actually
   calls today:

   ```go
   type Runner interface {
       RunProgram() error
       StopProgramAndCleanup() error
   }
   ```

   Verify the exact method set by grepping `server/strategies.go` for
   `goRun.` usages before finalizing — confirm whether `IsRunning` or other
   methods are also called from `server` and include only what's used.

2. Ensure `*gorun.GoRun` (returned by `gorun.New`) satisfies `Runner`
   with zero behavior changes — pure interface extraction.

3. Add a `FakeRunner` mock (same package/location convention decided in the
   `gobuild` plan — keep both mocks consistent):

   ```go
   type FakeRunner struct {
       RunErr        error
       StopErr       error
       RunCallCount  int
       StopCallCount int
   }

   func (f *FakeRunner) RunProgram() error {
       f.RunCallCount++
       return f.RunErr
   }
   func (f *FakeRunner) StopProgramAndCleanup() error {
       f.StopCallCount++
       return f.StopErr
   }
   ```

## Constraints (apply to all code in this plan)

- No hardcoded strings — any default identifiers in the mock must be named
  constants if repeated.
- Do not change `gorun.New`'s signature or `Config` struct — only add the
  interface and the mock.
- Must not break existing `gorun` tests (`RunProgram_test.go`,
  `StopProgram_test.go`, `IsRunning_test.go`, `cleanup_test.go`,
  `benchmark_test.go`) — these stay as real-subprocess integration tests,
  unchanged.
- Run `gotest ./...` inside `gorun/` after the change; all existing tests
  must still pass unmodified.

## Stages

| Stage | Description | Output |
|---|---|---|
| 1 | Grep `server/strategies.go` for `goRun.` usages to finalize interface method set | List of methods confirmed |
| 2 | Add `Runner` interface, verify `*GoRun` satisfies it (`var _ Runner = (*GoRun)(nil)`) | `gorun/runner.go` |
| 3 | Add `FakeRunner` mock | `gorun/mock/runner_mock.go` (or repo-conventional location, matching `gobuild`'s choice) |
| 4 | Run existing `gorun` test suite, confirm no regressions | Test output attached to PR |
