package types

// ImportOption describes options for a single import source.
type ImportOption struct {
	Params    *bool `json:"params,omitempty" yaml:"params,omitempty"`
	ConfigMap *bool `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	// Sessions can be either a boolean or a map[string]ImportSessionOption when parsed from config files.
	// Use interface{} here to allow flexible unmarshalling; callers may assert the concrete type.
	Sessions interface{} `json:"sessions,omitempty" yaml:"sessions,omitempty"`
}

// ImportSessionOption describes session-specific import settings.
type ImportSessionOption struct {
	Params    *bool `json:"params,omitempty" yaml:"params,omitempty"`
	ConfigMap *bool `json:"configMap,omitempty" yaml:"configMap,omitempty"`
}

// ParamOption defines metadata for a single parameter available to a command.
type ParamOption struct {
	Type        string      `json:"type" yaml:"type"` // "string" | "number" | "boolean" | "enum"
	Description *string     `json:"description,omitempty" yaml:"description,omitempty"`
	EnumValues  []string    `json:"enumValues,omitempty" yaml:"enumValues,omitempty"`
	Default     interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	Required    *bool       `json:"required,omitempty" yaml:"required,omitempty"`
}

// InformationalCommandSettings groups settings used when a command is informational.
type InformationalCommandSettings struct {
	Interval       *int     `json:"interval,omitempty" yaml:"interval,omitempty"`
	FileWatcher    []string `json:"fileWatcher,omitempty" yaml:"fileWatcher,omitempty"`
	CommandTrigger []string `json:"commandTrigger,omitempty" yaml:"commandTrigger,omitempty"`
	Order          *int     `json:"order,omitempty" yaml:"order,omitempty"`
}

// SessionConfig holds session-specific params and config map for a command.
type SessionConfig struct {
	Params    map[string]string `json:"params,omitempty" yaml:"params,omitempty"`
	ConfigMap map[string]string `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	Color     *string           `json:"color,omitempty" yaml:"color,omitempty"`
}

// InternalCommand represents a command definition.
// This mirrors the provided TypeScript InternalCommand interface and is used internally by the loader/runner.
type InternalCommand struct {
	// Populated by the system, not read directly from config files
	Name     string `json:"name" yaml:"name"`
	FullName string `json:"fullName" yaml:"fullName"`

	SkipTerminalAutofocus *bool `json:"skipTerminalAutofocus,omitempty" yaml:"skipTerminalAutofocus,omitempty"`

	ProtectionOnEnv string `json:"protectionOnEnv,omitempty" yaml:"protectionOnEnv,omitempty"`

	AutocloseCommand *bool `json:"autocloseCommand,omitempty" yaml:"autocloseCommand,omitempty"`

	Shell *string `json:"shell,omitempty" yaml:"shell,omitempty"` // expected: "bash" | "external" | "vscode" | "zsh"

	UseIntegratedShell     *bool    `json:"useIntegratedShell,omitempty" yaml:"useIntegratedShell,omitempty"`
	ExternalShellLocation  *string  `json:"externalShellLocation,omitempty" yaml:"externalShellLocation,omitempty"`
	ExternalShellArguments []string `json:"externalShellArguments,omitempty" yaml:"externalShellArguments,omitempty"`

	Template *string `json:"template,omitempty" yaml:"template,omitempty"`

	IsJSONOutput *bool `json:"isJSONOutput,omitempty" yaml:"isJSONOutput,omitempty"`
	HasTemplate  *bool `json:"hasTemplate,omitempty" yaml:"hasTemplate,omitempty"`

	IsInformational              *bool                         `json:"isInformational,omitempty" yaml:"isInformational,omitempty"`
	InformationalCommandSettings *InformationalCommandSettings `json:"informationalCommandSettings,omitempty" yaml:"informationalCommandSettings,omitempty"`

	Command string `json:"command,omitempty" yaml:"command,omitempty"`

	SubCommands []SubCommand `json:"subCommands,omitempty" yaml:"subCommands,omitempty"`

	KillCmd *string `json:"killCmd,omitempty" yaml:"killCmd,omitempty"`

	Pipe string `json:"pipe,omitempty" yaml:"pipe,omitempty"`

	DefaultFileName string `json:"defaultFileName,omitempty" yaml:"defaultFileName,omitempty"`
	TemplateName    string `json:"templateName,omitempty" yaml:"templateName,omitempty"`

	ImportFrom interface{} `json:"importFrom,omitempty" yaml:"importFrom,omitempty"` // string | []string | ImportOptions

	Extend *string `json:"extend,omitempty" yaml:"extend,omitempty"`

	Background *bool `json:"background,omitempty" yaml:"background,omitempty"`

	RequiredParams []string `json:"requiredParams,omitempty" yaml:"requiredParams,omitempty"`

	ForceEnv *bool `json:"forceEnv,omitempty" yaml:"forceEnv,omitempty"`

	Params map[string]string `json:"params,omitempty" yaml:"params,omitempty"`

	ParamOptions map[string]ParamOption `json:"paramOptions,omitempty" yaml:"paramOptions,omitempty"`

	ConfigMap map[string]string `json:"configMap,omitempty" yaml:"configMap,omitempty"`

	Sessions map[string]SessionConfig `json:"sessions,omitempty" yaml:"sessions,omitempty"`

	Color *string `json:"color,omitempty" yaml:"color,omitempty"`

	OriginalFilepath         *string `json:"originalFilepath,omitempty" yaml:"originalFilepath,omitempty"`
	OriginalTemplateFilepath *string `json:"originalTemplateFilepath,omitempty" yaml:"originalTemplateFilepath,omitempty"`

	EnableLog *bool `json:"enableLog,omitempty" yaml:"enableLog,omitempty"`

	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
}

// CommandConfigInternal defines the internal configuration structure for command management.
// Keep lightweight for now; expand as needed by the loader/runner.
type CommandConfigInternal struct {
	// Placeholder for runtime-only fields if needed in the future.
}
