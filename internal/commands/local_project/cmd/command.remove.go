package cmd

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func removeCommandCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <command-name>",
		Short: "Remove a command",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemoveCommand,
	}
}

func runRemoveCommand(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	commandName := args[0]
	index := findCommandIndex(localCfg, commandName)
	if index == -1 {
		return fmt.Errorf("command '%s' not found", commandName)
	}

	localCfg.Commands = append(localCfg.Commands[:index], localCfg.Commands[index+1:]...)
	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Command '%s' removed successfully.\n", commandName)
	return nil
}
