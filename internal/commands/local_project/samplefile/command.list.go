package samplefile

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <alias>",
		Short: "Get details about a sample file",
		Args:  cobra.ExactArgs(0),
		RunE:  runGet,
	}
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	env, err := localCfg.GetSelectedEnvironment()
	if err != nil {
		return fmt.Errorf("no environment selected. Use 'envtrack use' to select an environment")
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	formattedOutput, _ := formatter.Format(env.SampleFiles)
	fmt.Print(formattedOutput)

	return nil
}
