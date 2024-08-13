package cmd

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func addCommandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <command-name>",
		Short: "Add a new command",
		Args:  cobra.ExactArgs(1),
		RunE:  runAddCommand,
	}
	cmd.Flags().String("description", "", "Description of the command")
	cmd.Flags().String("command", "", "The command to execute")
	cmd.MarkFlagRequired("description")
	cmd.MarkFlagRequired("command")
	return cmd
}

func runAddCommand(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	commandName := args[0]
	description, _ := cmd.Flags().GetString("description")
	commandStr, _ := cmd.Flags().GetString("command")

	if findCommand(localCfg, commandName) != nil {
		return fmt.Errorf("command '%s' already exists", commandName)
	}

	newCommand := &config.LocalConfigCommand{
		Name:        commandName,
		Description: description,
		Command:     commandStr,
	}

	localCfg.Commands = append(localCfg.Commands, newCommand)
	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Command '%s' added successfully.\n", commandName)
	return nil
}
