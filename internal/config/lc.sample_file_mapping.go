package config

type LocalConfigSampleFileMapping struct {
	Variable          string `json:"variable"`                                             // Local LocalConfigSampleFile variable key
	VariableName      string `json:"variableName" yaml:"variableName,omitempty"`           // localConfig LocalConfigVariable name
	SecretName        string `json:"secretName" yaml:"secretName,omitempty"`               // localConfig LocalConfigSecret name
	LinkedFileName    string `json:"linkedFileName" yaml:"linkedFileName,omitempty"`       // env LocalConfigLinkedFile alias
	LinkedFileVarPath string `json:"linkedFileVarPath" yaml:"linkedFileVarPath,omitempty"` // env LocalConfigLinkedFile variable name
}
