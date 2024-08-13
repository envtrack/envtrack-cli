package cmd

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func listCommandCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available commands",
		Args:  cobra.NoArgs,
		RunE:  runListCommands,
	}
}

func runListCommands(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	output, err := formatter.Format(localCfg.Commands)
	if err != nil {
		return fmt.Errorf("error formatting output: %v", err)
	}

	fmt.Println(output)
	return nil
}
