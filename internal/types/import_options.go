package types

// ImportOptions defines flexible configuration for parameter and configuration map imports for different sessions.
type ImportOptions map[string]ImportSourceOptions

// ImportSourceOptions represents import options for a specific source.
type ImportSourceOptions struct {
	Params       *bool                        `json:"params,omitempty" yaml:"params,omitempty"`
	ConfigMap    *bool                        `json:"configMap,omitempty" yaml:"configMap,omitempty"`
	Sessions     map[string]SessionImportOpts `json:"sessions,omitempty" yaml:"sessions,omitempty"`         // Used if sessions is an object
	SessionsBool *bool                        `json:"sessionsBool,omitempty" yaml:"sessionsBool,omitempty"` // Used if sessions is a boolean
}

// SessionImportOpts represents session-specific import options.
type SessionImportOpts struct {
	Params    *bool `json:"params,omitempty" yaml:"params,omitempty"`
	ConfigMap *bool `json:"configMap,omitempty" yaml:"configMap,omitempty"`
}
