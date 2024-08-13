package cmd

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func getCommandCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <command-name>",
		Short: "Get details of a specific command",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetCommand,
	}
}

func runGetCommand(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	commandName := args[0]
	command := findCommand(localCfg, commandName)
	if command == nil {
		return fmt.Errorf("command '%s' not found", commandName)
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	output, err := formatter.Format(command)
	if err != nil {
		return fmt.Errorf("error formatting output: %v", err)
	}

	fmt.Println(output)
	return nil
}
