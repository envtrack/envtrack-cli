package cmd

import (
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func CommandCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command",
		Short: "Manage and execute predefined commands",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		getCommandCmd(),
		listCommandCmd(),
		addCommandCmd(),
		updateCommandCmd(),
		removeCommandCmd(),
		executeCommandCmd(),
	)

	return cmd
}

func findCommand(cfg *config.LocalConfigParams, name string) *config.LocalConfigCommand {
	for i := range cfg.Commands {
		if cfg.Commands[i].Name == name {
			return cfg.Commands[i]
		}
	}
	return nil
}

func findCommandIndex(cfg *config.LocalConfigParams, name string) int {
	for i := range cfg.Commands {
		if cfg.Commands[i].Name == name {
			return i
		}
	}
	return -1
}
