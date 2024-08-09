package varfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func linkVarfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Link an external variable file to the environment",
		RunE:  runLinkVarfile,
	}
	cmd.Flags().String("file", "", "Path to the external variable file (required)")
	cmd.Flags().StringP("alias", "a", "", "Alias for the linked file (required)")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("alias")
	return cmd
}

func runLinkVarfile(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	env, err := localCfg.GetSelectedEnvironment()
	if err != nil {
		return fmt.Errorf("no environment selected. Please use --environment flag or select an environment")
	}

	filePath, _ := cmd.Flags().GetString("file")
	alias, _ := cmd.Flags().GetString("alias")

	// Validate file exists
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error resolving file path: %v", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", absPath)
	}

	// Check if alias already exists
	for _, linkedFile := range env.LinkedFiles {
		if linkedFile.Alias == alias {
			return fmt.Errorf("alias '%s' is already in use", alias)
		}
	}

	// Add new linked file
	newLinkedFile := &config.LocalConfigLinkedFile{
		Path:  absPath,
		Alias: alias,
	}
	env.LinkedFiles = append(env.LinkedFiles, newLinkedFile)

	// Save updated configuration
	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Successfully linked file '%s' with alias '%s'.\n", absPath, alias)
	return nil
}
