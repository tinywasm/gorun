package gorun_test

import (
	"testing"

	"github.com/tinywasm/gorun"
	"github.com/tinywasm/gorun/mock"
)

func TestRunnerInterface(t *testing.T) {
	// Test that GoRun implements Runner
	var runner gorun.Runner
	runner = gorun.New(&gorun.Config{})
	if runner == nil {
		t.Fatal("GoRun should implement Runner interface")
	}

	// Test that FakeRunner implements Runner
	var fakeRunner gorun.Runner
	fakeRunner = &mock.FakeRunner{}
	if fakeRunner == nil {
		t.Fatal("FakeRunner should implement Runner interface")
	}
}

func TestFakeRunner(t *testing.T) {
	f := &mock.FakeRunner{}

	// Test RunProgram
	err := f.RunProgram()
	if err != f.RunErr {
		t.Errorf("expected error %v, got %v", f.RunErr, err)
	}
	if f.RunCallCount != 1 {
		t.Errorf("expected 1 call, got %d", f.RunCallCount)
	}

	// Test StopProgram
	err = f.StopProgram()
	if err != f.StopErr {
		t.Errorf("expected error %v, got %v", f.StopErr, err)
	}
	if f.StopCallCount != 1 {
		t.Errorf("expected 1 call, got %d", f.StopCallCount)
	}
}
