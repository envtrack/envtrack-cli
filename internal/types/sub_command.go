package types

// SubCommand represents a sub-command to be executed.
type SubCommand struct {
	// The command to be executed.
	CommandName string `json:"commandName" yaml:"commandName"`

	// Optional session in which the command should be executed.
	Session *string `json:"session,omitempty" yaml:"session,omitempty"`

	// Key-value pairs of parameters for the command.
	Params map[string]string `json:"params" yaml:"params"`
}
