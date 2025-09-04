package initialize

import (
	"encoding/json"
	"fmt"
	"github.com/envtrack/envtrack-cli/internal/commandconfig"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

// Flags to control output
var (
	hideCommandConfig   bool
	hideParsedConfigs   bool
	hideParsedTemplates bool
	hideParsedCommands  bool
)

// CmdDebugCommand returns a Cobra command that parses the commands configuration and
// prints the parsed structure for debugging/visualization.
func CmdDebugCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Parse and display the commands configuration (for debugging)",
		Run:   runCmdDebug,
	}

	// Add flags to control which parts of the output are shown. Provide shorthands.
	cmd.Flags().BoolVarP(&hideCommandConfig, "no-command-config", "c", false, "Do not include commandConfig in output")
	cmd.Flags().BoolVarP(&hideParsedConfigs, "no-parsed-configs", "p", false, "Do not include parsedConfigs in output")
	cmd.Flags().BoolVarP(&hideParsedTemplates, "no-parsed-templates", "t", false, "Do not include parsedTemplates in output")
	cmd.Flags().BoolVarP(&hideParsedCommands, "no-parsed-commands", "m", false, "Do not include parsedCommands in output")

	return cmd
}

func runCmdDebug(_ *cobra.Command, _ []string) {
	localCfgParams, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfgParams == nil {
		fmt.Printf("Error loading local config: %v\n", err)
		return
	}

	if localCfgParams.CommandsConfiguration == nil {
		fmt.Println("No commands configuration set in local config.")
		return
	}

	commandsConfiguration, err := commandconfig.LoadFromFile(*localCfgParams.CommandsConfiguration.ExternalConfigPath)
	if err != nil {
		fmt.Printf("Error loading commands configuration from file: %v\n", err)
		return
	}

	// Otherwise, if the config is embedded in the local config, print it and parse any referenced configs.
	parsedConfigs, pErr := commandconfig.ParseConfigsFromCommandConfig(commandsConfiguration)
	if pErr != nil {
		fmt.Printf("Warning: error parsing referenced configs: %v\n", pErr)
	}

	parsedTemplates, tErr := commandconfig.ParseTemplatesFromCommandConfig(commandsConfiguration)
	if tErr != nil {
		fmt.Printf("Warning: error parsing templates: %v\n", tErr)
	}

	// Parse typed InternalCommand objects
	parsedCommands, pcErr := commandconfig.ParseInternalCommandsFromCommandConfig(commandsConfiguration)
	if pcErr != nil {
		fmt.Printf("Warning: error parsing commands: %v\n", pcErr)
	}

	outObj := map[string]interface{}{}
	if !hideCommandConfig {
		outObj["commandConfig"] = commandsConfiguration
	}
	if !hideParsedConfigs {
		outObj["parsedConfigs"] = parsedConfigs
	}
	if !hideParsedTemplates {
		outObj["parsedTemplates"] = parsedTemplates
	}
	if !hideParsedCommands {
		outObj["parsedCommands"] = parsedCommands
	}

	// If nothing was selected, inform the user instead of printing an empty object.
	if len(outObj) == 0 {
		fmt.Println("No output selected (commandConfig, parsedConfigs, parsedTemplates and parsedCommands are hidden). Use flags to include output.")
		return
	}

	out, jerr := json.MarshalIndent(outObj, "", "  ")
	if jerr != nil {
		fmt.Printf("Error serializing local commands configuration: %v\n", jerr)
		return
	}

	fmt.Println(string(out))
}

// Prevent unused function warning when static analysis doesn't detect cross-package usage.
var _ = CmdDebugCommand
