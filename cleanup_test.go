package gorun

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestKillAllOnStop_Disabled(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	_, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
		KillAllOnStop:   false, // Disabled
	}

	gr := New(config)

	// Start and stop program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	err = gr.StopProgram()
	if err != nil && !strings.Contains(err.Error(), "no child processes") {
		t.Errorf("StopProgram() failed with unexpected error: %v", err)
	}

	// Should work normally (this test mainly ensures no regressions)
	if gr.IsRunning() {
		t.Error("Program should not be running after stop")
	}
}

func TestKillAllOnStop_Enabled(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	_, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
		KillAllOnStop:   true, // Enabled
	}

	gr := New(config)

	// Start and stop program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	err = gr.StopProgram()
	if err != nil && !strings.Contains(err.Error(), "no child processes") {
		t.Errorf("StopProgram() failed with unexpected error: %v", err)
	}

	// Should work normally (this test mainly ensures no regressions)
	if gr.IsRunning() {
		t.Error("Program should not be running after stop")
	}
}

func TestStopProgramAndCleanup(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	_, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
		KillAllOnStop:   false,
	}

	gr := New(config)

	// Start program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Test explicit cleanup call with killAll=true
	err = gr.StopProgramAndCleanup(true)
	if err != nil && !strings.Contains(err.Error(), "no child processes") {
		t.Errorf("StopProgramAndCleanup(true) failed with unexpected error: %v", err)
	}

	if gr.IsRunning() {
		t.Error("Program should not be running after cleanup")
	}
}

func TestStopProgramAndCleanup_NoCleanup(t *testing.T) {
	// Build test program
	execPath := buildTestProgram(t, "simple_program")
	defer os.Remove(execPath)

	exitChan := make(chan bool)
	_, logger := createTestLogger()

	config := &Config{
		ExecProgramPath: execPath,
		RunArguments:    func() []string { return []string{} },
		ExitChan:        exitChan,
		Logger:          logger,
		KillAllOnStop:   false,
	}

	gr := New(config)

	// Start program
	err := gr.RunProgram()
	if err != nil {
		t.Fatalf("RunProgram() failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Test explicit cleanup call with killAll=false
	err = gr.StopProgramAndCleanup(false)
	if err != nil && !strings.Contains(err.Error(), "no child processes") {
		t.Errorf("StopProgramAndCleanup(false) failed with unexpected error: %v", err)
	}

	if gr.IsRunning() {
		t.Error("Program should not be running after cleanup")
	}
}

func TestKillAllByName_NonExistent(t *testing.T) {
	// Test killing a non-existent program - should not error
	err := KillAllByName("nonexistent_program_12345", false)
	if err != nil {
		t.Errorf("KillAllByName should not error for non-existent program: %v", err)
	}
}
