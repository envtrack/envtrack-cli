package cmd

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func updateCommandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <command-name>",
		Short: "Update an existing command",
		Args:  cobra.ExactArgs(1),
		RunE:  runUpdateCommand,
	}
	cmd.Flags().String("description", "", "New description of the command")
	cmd.Flags().String("command", "", "New command to execute")
	return cmd
}

func runUpdateCommand(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	commandName := args[0]
	command := findCommand(localCfg, commandName)
	if command == nil {
		return fmt.Errorf("command '%s' not found", commandName)
	}

	description, _ := cmd.Flags().GetString("description")
	commandStr, _ := cmd.Flags().GetString("command")

	if description != "" {
		command.Description = description
	}
	if commandStr != "" {
		command.Command = commandStr
	}

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Command '%s' updated successfully.\n", commandName)
	return nil
}
