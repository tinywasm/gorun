package gorun

// Runner is an interface for running and stopping programs.
type Runner interface {
	RunProgram() error
	StopProgram() error
}
