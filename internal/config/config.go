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

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName(configFileName)
	viper.SetConfigFile(filepath.Join(configDir, configFileName+".yaml"))

	viper.SetDefault("api_endpoint", defaultAPIEndpoint)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			viper.WriteConfig()
		}
	}
}

func SetConfig(key, value string) error {
	viper.Set(key, value)
	return viper.WriteConfig()
}

func GetConfig(key string) string {
	return viper.GetString(key)
}

func SetAuthToken(token string) error {
	return keyring.Set(appName, authTokenKey, token)
}

func GetAuthToken() (string, error) {
	return keyring.Get(appName, authTokenKey)
}

const (
	defaultFormatKey = "default_format"
	outputFormatKey  = "output_format"
)

func SetDefaultFormat(format string) error {
	return SetConfig(defaultFormatKey, format)
}

func GetDefaultFormat() string {
	return GetConfig(defaultFormatKey)
}

func WithOutputFormat(ctx context.Context, format string) context.Context {
	return context.WithValue(ctx, outputFormatKey, format)
}

func GetOutputFormat(ctx context.Context) string {
	if format, ok := ctx.Value(outputFormatKey).(string); ok {
		return format
	}
	return "json" // Default to JSON if not set
}
