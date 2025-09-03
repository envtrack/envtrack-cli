package config

import (
	"fmt"
	"github.com/envtrack/envtrack-cli/internal/types"
)

type LocalConfigParams struct {
	Organization          *LocalConfigOrganization  `json:"organization" yaml:"organization"`
	Project               *LocalConfigProject       `json:"project" yaml:"project"`
	Environments          []*LocalConfigEnvironment `json:"environments" yaml:"environments"`
	SelectedEnv           string                    `json:"selectedEnv" yaml:"selectedEnv"`
	CommandsConfiguration *types.CommandConfig      `json:"commandsConfiguration" yaml:"commandsConfiguration"`
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
