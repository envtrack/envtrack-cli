package config

type LocalConfigSampleFile struct {
	Alias     string                          `json:"alias"`
	Path      string                          `json:"path"`
	Variables []string                        `json:"variables"`
	Mapping   []*LocalConfigSampleFileMapping `json:"mapping"`
}
