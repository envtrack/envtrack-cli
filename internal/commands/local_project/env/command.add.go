package env

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func addEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "Add a new environment to the local configuration",
		Run:     runAddEnvironment,
		PreRunE: validateAddEnvironmentFlags,
	}

	cmd.Flags().StringP("name", "n", "", "Name of the environment")
	cmd.Flags().StringP("shortname", "s", "", "Short name of the environment")
	cmd.Flags().Bool("local", false, "Add environment locally without syncing to server")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("shortname")

	return cmd
}

func validateAddEnvironmentFlags(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	shortName, _ := cmd.Flags().GetString("shortname")

	if name == "" || shortName == "" {
		return fmt.Errorf("both name and shortname are required")
	}

	return nil
}

func runAddEnvironment(cmd *cobra.Command, args []string) {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		fmt.Println("No local context found. Use 'envtrack init' to initialize a local project.")
		return
	}

	name, _ := cmd.Flags().GetString("name")
	shortName, _ := cmd.Flags().GetString("shortname")
	local, _ := cmd.Flags().GetBool("local")

	// Check if environment already exists
	for _, env := range localCfg.Environments {
		if env.Name == name || env.ShortName == shortName {
			fmt.Printf("Environment with name '%s' or shortname '%s' already exists.\n", name, shortName)
			return
		}
	}

	var newEnv config.LocalConfigEnvironment

	if !local {
		// Sync with server
		token, _ := config.GlobalConf.GetAuthToken()
		client := api.NewClient(token)
		serverEnv, err := client.CreateEnvironment(localCfg.Organization.ID, localCfg.Project.ID, name, shortName)
		if err != nil {
			fmt.Printf("Error creating environment on server: %v\n", err)
			return
		}
		newEnv = config.LocalConfigEnvironment{
			ID:        serverEnv.ID,
			Name:      serverEnv.Name,
			ShortName: serverEnv.ShortName,
		}
	} else {
		// Local only
		newEnv = config.LocalConfigEnvironment{
			ID:        "", // Leave empty for local-only environments
			Name:      name,
			ShortName: shortName,
		}
	}

	localCfg.Environments = append(localCfg.Environments, newEnv)

	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		fmt.Printf("Error saving local configuration: %v\n", err)
		return
	}

	fmt.Printf("Environment '%s' (%s) added successfully.\n", name, shortName)
}
