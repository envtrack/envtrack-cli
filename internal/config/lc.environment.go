package config


type LocalConfigEnvironment struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	ShortName   string                   `json:"shortName" yaml:"shortName"`
	Variables   []*LocalConfigVariable   `json:"variables"`
	Secrets     []*LocalConfigSecret     `json:"secrets"`
	LinkedFiles []*LocalConfigLinkedFile `json:"linkedFiles"`
	SampleFiles []*LocalConfigSampleFile `json:"sampleFiles"`
	IsSelected  bool                     `json:"isSelected"`
}

func (e *LocalConfigEnvironment) GetVariable(name string) *LocalConfigVariable {
	for _, v := range e.Variables {
		if v.Name == name {
			return v
		}
	}
	return nil
}
