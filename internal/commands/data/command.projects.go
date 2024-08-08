package data

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func ProjectsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"prj"},
		GroupID: "data",
		Short:   "List projects for a given organization",
		Run:     common.RequireAuth(runProjects),
	}
	cmd.Flags().StringP("organization", "o", "", "ID or shortname of the organization")
	cmd.MarkFlagRequired("organization")
	return cmd
}

func runProjects(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Flags().GetString("organization")

	token, _ := config.GlobalConf.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetProjects(orgID)
	if err != nil {
		fmt.Printf("Error fetching organizations: %v\n", err)
		return
	}

	formatter, err := common.GetFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(orgs)
	fmt.Print(formattedOutput)
}
