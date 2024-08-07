package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

const (
	appName            = "envtrack-cli"
	configFileName     = "config"
	authTokenKey       = "auth_token"
	defaultAPIEndpoint = "https://europe-west1-envtrack-2fd23.cloudfunctions.net"
)

var (
	configDir string
)

func init() {
	GlobalConf = &globalConfStruct{
		v: viper.New(),
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	configDir = filepath.Join(homeDir, "."+appName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	GlobalConf.v.AddConfigPath(configDir)
	GlobalConf.v.SetConfigType("yaml")
	GlobalConf.v.SetConfigName(configFileName)
	GlobalConf.v.SetConfigFile(filepath.Join(configDir, configFileName+".yaml"))

	GlobalConf.v.SetDefault("api_endpoint", defaultAPIEndpoint)

	if err := GlobalConf.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			GlobalConf.v.WriteConfig()
		}
	}
}

type globalConfStruct struct {
	v *viper.Viper
}

var GlobalConf = &globalConfStruct{}

func (c *globalConfStruct) Set(key, value string) error {
	viper.Set(key, value)
	return viper.WriteConfig()
}

func (c *globalConfStruct) Get(key string) string {
	return viper.GetString(key)
}

func (c *globalConfStruct) SetAuthToken(token string) error {
	return keyring.Set(appName, authTokenKey, token)
}

func (c *globalConfStruct) GetAuthToken() (string, error) {
	return keyring.Get(appName, authTokenKey)
}

type envtrackContextKey string

const (
	defaultFormatKey envtrackContextKey = "default_format"
	outputFormatKey  envtrackContextKey = "output_format"
)

func (c *globalConfStruct) SetDefaultFormat(format string) error {
	return c.Set(string(defaultFormatKey), format)
}

func (c *globalConfStruct) GetDefaultFormat() string {
	return c.Get(string(defaultFormatKey))
}

func (c *globalConfStruct) WithOutputFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, outputFormatKey, format)
}

func (c *globalConfStruct) GetOutputFormat(ctx context.Context) string {
	if format, ok := ctx.Value(outputFormatKey).(string); ok {
		return format
	}
	return "json" // Default to JSON if not set
}
