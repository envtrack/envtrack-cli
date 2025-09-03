package types

// CommandConfigConfig defines the configuration for command execution settings.
type CommandConfigConfig struct {
	// The location where the command should be executed.
	Location string `json:"location" yaml:"location"`

	// A key-value pair object containing variables for the command.
	Variables map[string]interface{} `json:"variables" yaml:"variables"`

	// Optional path to an external config file which defines additional variables or settings.
	// This path may be relative to the command config's root path.
	ConfigPath *string `json:"configPath,omitempty" yaml:"configPath,omitempty"`
}
