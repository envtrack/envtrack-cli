package config

type LocalConfigOrganization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName" yaml:"shortName"`
}
