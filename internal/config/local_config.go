package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	localConfigName = ".envtrack"
)

type localConfStruct struct {
	v *viper.Viper
}

var LocalConf localConfStruct

type LocalConfigParams struct {
	Organization LocalConfigOrganization `json:"organization"`
	Project      LocalConfigProject      `json:"project"`
}

type LocalConfigOrganization struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName" yaml:"shortName"`
}

type LocalConfigProject struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName" yaml:"shortName"`
}

func init() {
	LocalConf = localConfStruct{
		v: viper.New(),
	}
	LocalConf.v.SetConfigName(localConfigName)
	LocalConf.v.SetConfigType("yaml")
	LocalConf.v.AddConfigPath(".")
	LocalConf.v.SetConfigFile(localConfigName + ".yaml")

	if err := LocalConf.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			LocalConf.v.WriteConfig()
		}
	}
}

func (c *localConfStruct) Set(key, value string) error {
	viper.Set(key, value)
	return viper.WriteConfig()
}

func (c *localConfStruct) Get(key string) string {
	return viper.GetString(key)
}

func (c *localConfStruct) SaveLocalConfig(config LocalConfigParams) error {
	c.v.Set("organization.id", config.Organization.ID)
	c.v.Set("project.id", config.Project.ID)
	c.v.Set("organization.name", config.Organization.Name)
	c.v.Set("project.name", config.Project.Name)
	c.v.Set("organization.shortName", config.Organization.ShortName)
	c.v.Set("project.shortName", config.Project.ShortName)

	err := c.v.WriteConfig()
	if err != nil {
		return fmt.Errorf("error writing local config: %w", err)
	}

	return c.addToGitignore(localConfigName + ".yaml")
}

func (c *localConfStruct) GetLocalConfig() (*LocalConfigParams, error) {
	err := c.v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading local config: %w", err)
	}

	var config LocalConfigParams
	err = c.v.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling local config: %w", err)
	}

	return &config, nil
}

func (c *localConfStruct) addToGitignore(filename string) error {
	gitignorePath := ".gitignore"

	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil // .gitignore doesn't exist, so we don't need to modify it
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("error reading .gitignore: %w", err)
	}

	if !strings.Contains(string(content), filename) {
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening .gitignore: %w", err)
		}
		defer f.Close()

		_, err = f.WriteString("\n\n#EnvTrack config\n" + filename + "\n")
		if err != nil {
			return fmt.Errorf("error writing to .gitignore: %w", err)
		}
	}

	return nil
}
