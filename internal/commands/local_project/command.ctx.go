package local_project

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/commands/local_project/env"
	"github.com/envtrack/envtrack-cli/internal/commands/local_project/initialize"
	"github.com/envtrack/envtrack-cli/internal/commands/local_project/secrets"
	"github.com/envtrack/envtrack-cli/internal/commands/local_project/variables"
	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func LocalContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Aliases: []string{"ctx"},
		GroupID: "local_project",
		Short:   "Display or manage local EnvTrack context",
		Run:     runLocalContext,
	}
	cmd.AddCommand(initialize.InitCommand())
	cmd.AddCommand(env.LocalEnvCommand())
	cmd.AddCommand(variables.LocalVariablesCommand())
	cmd.AddCommand(secrets.SecretsCommand())
	return cmd
}

func runLocalContext(cmd *cobra.Command, args []string) {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil || localCfg == nil {
		fmt.Println("No local context found. Use 'envtrack ctx init' to initialize a local project.")
		return
	}

	if localCfg.Organization.ID == "" || localCfg.Project.ID == "" {
		fmt.Println("Local context is incomplete. Use 'envtrack ctx init' to initialize a local project.")
		return
	}

	output := map[string]string{
		"Organization": fmt.Sprintf("%s (%s)", localCfg.Organization.Name, localCfg.Organization.ShortName),
		"Project":      fmt.Sprintf("%s (%s)", localCfg.Project.Name, localCfg.Project.ShortName),
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(output)
	fmt.Print(formattedOutput)
}
