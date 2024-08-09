package samplefile

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func getCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <alias>",
		Short: "Get details about a sample file",
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
	}
	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
	alias := args[0]

	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	env, err := localCfg.GetSelectedEnvironment()
	if err != nil {
		return fmt.Errorf("no environment selected. Use 'envtrack use' to select an environment")
	}

	var sampleFile *config.LocalConfigSampleFile
	for _, sf := range env.SampleFiles {
		if sf.Alias == alias {
			sampleFile = sf
			break
		}
	}

	if sampleFile == nil {
		return fmt.Errorf("sample file with alias '%s' not found", alias)
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		return fmt.Errorf("error getting formatter: %v", err)
	}

	formattedOutput, _ := formatter.Format(sampleFile)
	fmt.Print(formattedOutput)

	return nil
}
