package containeragentdto

import "context"

// ConsoleSize holds the dimensions of a console terminal
type ConsoleSize struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

// ExecConfig represents the initial configuration for command execution in a container
type ExecConfig struct {
	ContainerID string       `json:"containerId"`
	Cmd         []string     `json:"cmd"`
	Env         []string     `json:"env"`
	WorkingDir  string       `json:"workingDir"`
	Tty         bool         `json:"tty"`
	ConsoleSize *ConsoleSize `json:"consoleSize"`
}

// ResizeOptions represents console resize dimensions
type ResizeOptions struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

// ExecInput represents incoming stream events from the client
type ExecInput struct {
	Config *ExecConfig    `json:"config"`
	Stdin  []byte         `json:"stdin"`
	Resize *ResizeOptions `json:"resize"`
}

// ExecOutput represents outgoing stream events to the client
type ExecOutput struct {
	Stdout   []byte `json:"stdout"`
	Stderr   []byte `json:"stderr"`
	ExitCode *int32 `json:"exitCode"`
}

// ExecStream represents the bidirectional stream interface for container execution
type ExecStream interface {
	Context() context.Context
	Recv() (*ExecInput, error)
	Send(*ExecOutput) error
}
