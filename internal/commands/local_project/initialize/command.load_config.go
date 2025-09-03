package initialize

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/envtrack/envtrack-cli/internal/commandconfig"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/envtrack/envtrack-cli/internal/types"
	"github.com/spf13/cobra"
)

func LoadConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load-config",
		Short: "Load a command config file into the local configuration",
		Run:   runLoadConfig,
	}
	cmd.Flags().StringP("configPath", "c", "", "Path to the command config file (JSON or YAML)")
	err := cmd.MarkFlagRequired("configPath")
	if err != nil {
		fmt.Printf("Error setting up command: %v\n", err)
		return nil
	}
	return cmd
}

func runLoadConfig(cmd *cobra.Command, _ []string) {
	configPath, _ := cmd.Flags().GetString("configPath")

	// Validate the provided config file using the new commandconfig library
	if err := commandconfig.ValidateConfigFile(configPath); err != nil {
		fmt.Printf("Invalid command config file: %v\n", err)
		return
	}

	// Compute a path relative to the local config (assumed to be the current working directory)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	relPath, err := filepath.Rel(cwd, configPath)
	if err != nil {
		// fallback to absolute path if relative cannot be computed
		abs, aerr := filepath.Abs(configPath)
		if aerr != nil {
			fmt.Printf("Error determining path for config: %v\n", aerr)
			return
		}
		relPath = abs
	}

	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfg == nil {
		fmt.Printf("Error loading local config: %v\n", err)
		return
	}

	// Only store the external path in the local config; parsing/working with the config
	// will be done via the commandconfig library when needed.
	var commandConfig types.CommandConfig
	commandConfig.ExternalConfigPath = &relPath

	localCfg.CommandsConfiguration = &commandConfig
	if err := config.LocalConf.SaveLocalConfig(*localCfg); err != nil {
		fmt.Printf("Error saving local config: %v\n", err)
		return
	}

	fmt.Println("Command configuration external path saved successfully.")
}
