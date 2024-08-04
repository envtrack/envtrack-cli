package commands

import (
	"context"
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/envtrack/envtrack-cli/internal/output"
	"github.com/spf13/cobra"
)

func getFormatter(ctx context.Context) (output.Formatter, error) {
	format := config.GetOutputFormat(ctx)
	formatter, err := output.GetFormatter(format)
	if err != nil {
		return nil, err
	}

	return formatter, nil

}

func runOrganizations(cmd *cobra.Command, args []string) {
	token, _ := config.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetOrganizations()
	if err != nil {
		fmt.Printf("Error fetching organizations: %v\n", err)
		return
	}

	formatter, err := getFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(orgs)
	fmt.Printf(formattedOutput)
}

func runProjects(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Flags().GetString("organizationId")

	token, _ := config.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetProjects(orgID)
	if err != nil {
		fmt.Printf("Error fetching organizations: %v\n", err)
		return
	}

	formatter, err := getFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(orgs)
	fmt.Printf(formattedOutput)
}

func runEnvironments(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Flags().GetString("organizationId")
	projID, _ := cmd.Flags().GetString("projectId")

	token, _ := config.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetEnvironments(orgID, projID)
	if err != nil {
		fmt.Printf("Error fetching organizations: %v\n", err)
		return
	}

	formatter, err := getFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(orgs)
	fmt.Printf(formattedOutput)
}

func runVariables(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Flags().GetString("organizationId")
	projID, _ := cmd.Flags().GetString("projectId")
	envID, _ := cmd.Flags().GetString("environmentId")

	token, _ := config.GetAuthToken() // Error already checked in requireAuth
	client := api.NewClient(token)

	orgs, err := client.GetVariables(orgID, projID, envID)
	if err != nil {
		fmt.Printf("Error fetching organizations: %v\n", err)
		return
	}

	formatter, err := getFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(orgs)
	fmt.Printf(formattedOutput)
}
