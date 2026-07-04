package mock

import "github.com/tinywasm/gorun"

// FakeRunner is a mock implementation of the gorun.Runner interface.
type FakeRunner struct {
	RunErr        error
	StopErr       error
	RunCallCount  int
	StopCallCount int
}

// RunProgram simulates running a program.
func (f *FakeRunner) RunProgram() error {
	f.RunCallCount++
	return f.RunErr
}

// StopProgram simulates stopping a program.
func (f *FakeRunner) StopProgram() error {
	f.StopCallCount++
	return f.StopErr
}

// Ensure FakeRunner satisfies the gorun.Runner interface at compile time.
var _ gorun.Runner = (*FakeRunner)(nil)
