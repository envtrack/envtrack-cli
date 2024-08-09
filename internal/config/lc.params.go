package config

import "fmt"

type LocalConfigParams struct {
	Organization *LocalConfigOrganization  `json:"organization"`
	Project      *LocalConfigProject       `json:"project"`
	Environments []*LocalConfigEnvironment `json:"environments"`
	SelectedEnv  string                    `json:"selectedEnv"`
}

func (lc *LocalConfigParams) GetSelectedEnvironment() (*LocalConfigEnvironment, error) {
	return lc.GetEnvironment(lc.SelectedEnv)
}

func (lc *LocalConfigParams) GetEnvironment(shortNameOrID string) (*LocalConfigEnvironment, error) {
	for _, env := range lc.Environments {
		if env.ID == shortNameOrID || env.ShortName == shortNameOrID {
			return env, nil
		}
	}

	return nil, fmt.Errorf("environment not found")
}
