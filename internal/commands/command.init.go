package commands

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func initCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		GroupID: "local_project",
		Short:   "Initialize a local EnvTrack project",
		Run:     runInit,
		PreRunE: validateInitFlags,
	}
	cmd.Flags().StringP("organization", "o", "", "ID or shortname of the organization")
	cmd.Flags().StringP("project", "p", "", "ID or shortname of the project")
	cmd.Flags().String("projectName", "", "Name of the project (to be used with local, disregarded if project is loaded from server)")
	cmd.Flags().String("projectShortName", "", "Shortname of the project (to be used with local, disregarded if project is loaded from server)")
	cmd.Flags().String("organizationName", "", "Name of the organization (to be used with local, disregarded if project is loaded from server)")
	cmd.Flags().String("organizationShortName", "", "Shortname of the organization (to be used with local, disregarded if project is loaded from server)")
	cmd.Flags().BoolP("local", "l", false, "Initialize a local project without getting data from the server")
	cmd.MarkFlagRequired("organization")
	cmd.MarkFlagRequired("project")
	return cmd
}

func runInit(cmd *cobra.Command, args []string) {
	orgID, _ := cmd.Flags().GetString("organization")
	projID, _ := cmd.Flags().GetString("project")
	orgName, _ := cmd.Flags().GetString("organizationName")
	orgShortName, _ := cmd.Flags().GetString("organizationShortName")
	projName, _ := cmd.Flags().GetString("projectName")
	projShortName, _ := cmd.Flags().GetString("projectShortName")

	local, _ := cmd.Flags().GetBool("local")
	if !local {
		token, _ := config.GlobalConf.GetAuthToken() // Error already checked in requireAuth
		client := api.NewClient(token)

		project, organization, err := client.GetProjectWithOrganization(orgID, projID)
		if err != nil {
			fmt.Printf("Error fetching project: %v\n", err)
			return
		}

		orgID = organization.ID
		projID = project.ID
		orgName = organization.Name
		orgShortName = organization.ShortName
		projName = project.Name
		projShortName = project.ShortName
	}

	localCfg, err := config.LocalConf.GetLocalConfig()
	if err == nil && localCfg.Project.ID != "" && localCfg.Organization.ID != "" {
		fmt.Printf("Local project already initialized for \"%s\"/\"%s\"\n", localCfg.Organization.ShortName, localCfg.Project.ShortName)
		return
	}

	err = config.LocalConf.SaveLocalConfig(config.LocalConfigParams{
		Project: config.LocalConfigProject{
			ID:        projID,
			Name:      projName,
			ShortName: projShortName,
		},
		Organization: config.LocalConfigOrganization{
			ID:        orgID,
			Name:      orgName,
			ShortName: orgShortName,
		},
	})
	if err != nil {
		fmt.Printf("Error initializing local project: %v\n", err)
		return
	}

	fmt.Println("Local project initialized successfully.")
}

func validateInitFlags(cmd *cobra.Command, args []string) error {
	local, _ := cmd.Flags().GetBool("local")

	if local {
		requiredFlags := []string{"organizationName", "organizationShortName", "projectName", "projectShortName"}
		for _, flag := range requiredFlags {
			if value, _ := cmd.Flags().GetString(flag); value == "" {
				return fmt.Errorf("required flag \"%s\" not set", flag)
			}
		}
	} else {
		// should be empty
		emptyFlags := []string{"organizationName", "organizationShortName", "projectName", "projectShortName"}
		for _, flag := range emptyFlags {
			if value, _ := cmd.Flags().GetString(flag); value != "" {
				return fmt.Errorf("flag \"%s\" should not be set when initializing a project from the server", flag)
			}
		}
	}

	return nil
}
