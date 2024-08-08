package data

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/common"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func OrganizationsCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"org"},
		GroupID: "data",
		Short:   "List organizations",
		Run:     common.RequireAuth(runOrganizations),
	}
}

func runOrganizations(cmd *cobra.Command, args []string) {
	token, _ := config.GlobalConf.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetOrganizations()
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
