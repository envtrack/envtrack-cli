package cmd

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/envtrack/envtrack-cli/internal/exec"
	"github.com/spf13/cobra"
)

func executeCommandCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "execute <command-name>",
		Short: "Execute a predefined command",
		Args:  cobra.ExactArgs(1),
		RunE:  runExecuteCommand,
	}
}

func runExecuteCommand(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	commandName := args[0]
	command := findCommand(localCfg, commandName)
	if command == nil {
		return fmt.Errorf("command '%s' not found", commandName)
	}

	fmt.Printf("Executing command: %s\n%s\n", command.Description, command.Command)

	manager := exec.NewCommandManager("test_log")
	manager.AddCommand(&exec.Command{
		Name:       command.Name,
		Command:    command.Command,
		Background: false,
		// Identifier: command.Name,
	})

	if err := manager.ExecuteCommand(command.Name); err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	return nil
}
