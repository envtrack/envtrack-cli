package variableparser

type VariableParserParams struct {
	Organization *VariableParserOrganization `json:"organization"`
	Project      *VariableParserProject      `json:"project"`
	Env          *VariableParserEnvironment  `json:"env"`
}

type VariableParserOrganization struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName" yaml:"shortName"`
}

type VariableParserProject struct {
	Name      string `json:"name"`
	ShortName string `json:"shortName" yaml:"shortName"`
}

type VariableParserEnvironment struct {
	Name      string                                        `json:"name"`
	ShortName string                                        `json:"shortName" yaml:"shortName"`
	Vars      map[string]string                             `json:"vars"`
	Secrets   map[string]string                             `json:"secrets"`
	Varfile   map[string]*VariableParserLocalFilesVariables `json:"varfile"`
	Mappings  map[string]string                             `json:"mappings"`
}

type VariableParserLocalFilesVariables struct {
}
