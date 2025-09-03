package types

// CommandConfig represents the complete configuration for a command, combining VS Code-specific and internal command configurations.
type CommandConfig struct {
	FilePath *string `json:"filePath,omitempty" yaml:"filePath,omitempty"`

	// Path to an external configuration file, if applicable
	ExternalConfigPath *string `json:"externalConfigPath,omitempty" yaml:"externalConfigPath,omitempty"`

	// Array of paths to command template files
	CommandTemplatesPaths []string `json:"commandTemplatesPaths,omitempty" yaml:"commandTemplatesPaths,omitempty"`

	// Array of paths to folders containing command definitions
	CommandFolderPaths []string `json:"commandFolderPaths,omitempty" yaml:"commandFolderPaths,omitempty"`

	// A record of all available commands, where the key is the command name and the value is the InternalCommand object
	Commands map[string]InternalCommand `json:"commands" yaml:"commands"`

	// List of command names that should be monitored or observed
	ObservableCommands []string `json:"observableCommands" yaml:"observableCommands"`

	// Configuration for automatic command execution on startup. Keys are command names, values are session names or nil.
	Autostart map[string]*string `json:"autostart,omitempty" yaml:"autostart,omitempty"`

	// Name of the default session to be used when none is specified
	DefaultSession *string `json:"defaultSession,omitempty" yaml:"defaultSession,omitempty"`

	// If true, the default session will be automatically selected when the application starts
	AutoSelectDefaultSession *bool `json:"autoSelectDefaultSession,omitempty" yaml:"autoSelectDefaultSession,omitempty"`

	// A record of configuration settings for different commands or contexts. The key is a unique identifier, and the value is a CommandConfigConfig object.
	Configs map[string]CommandConfigConfig `json:"configs" yaml:"configs"`

	// A map of key-value pairs for custom path aliases
	Paths map[string]string `json:"paths,omitempty" yaml:"paths,omitempty"`

	// A record of template commands, built dynamically rather than imported from config.
	// Each key is a template name, and the value is an InternalCommand object.
	Templates map[string]InternalCommand `json:"templates,omitempty" yaml:"templates,omitempty"`

	// The root dir path of the config file.
	// This can be used for resolving relative paths or determining the context of the command.
	RootPath *string `json:"rootPath,omitempty" yaml:"rootPath,omitempty"`

	// The path to the project or workspace where the command is being executed.
	ProjectRootPath *string `json:"projectRootPath,omitempty" yaml:"projectRootPath,omitempty"`

	// The file path to the configuration file that defines this command.
	ConfigPath *string `json:"configPath,omitempty" yaml:"configPath,omitempty"`
}
