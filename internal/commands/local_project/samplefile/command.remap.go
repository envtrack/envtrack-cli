package samplefile

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func remapCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remap",
		Short: "Remap variables for an existing sample file",
		RunE:  runRemap,
	}
	cmd.Flags().StringP("alias", "a", "", "Alias of the sample file to remap (required)")
	cmd.MarkFlagRequired("alias")
	return cmd
}

func runRemap(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	alias, _ := cmd.Flags().GetString("alias")

	env, err := localCfg.GetSelectedEnvironment()
	if err != nil {
		return fmt.Errorf("no environment selected. Use 'envtrack use' to select an environment")
	}

	var sampleFile *config.LocalConfigSampleFile
	for i, sf := range env.SampleFiles {
		if sf.Alias == alias {
			sampleFile = env.SampleFiles[i]
			break
		}
	}

	if sampleFile == nil {
		return fmt.Errorf("sample file with alias '%s' not found", alias)
	}

	// Re-create the mapping
	varMapping, err := createMapping(sampleFile.Path)
	if err != nil {
		return err
	}

	// Clear existing mapping and create new one
	sampleFile.Mapping = []*config.LocalConfigSampleFileMapping{}
	sampleFile.Variables = []string{}
	for _, key := range varMapping {
		sampleFile.Variables = append(sampleFile.Variables, key)
		// sampleFile.Mapping = append(sampleFile.Mapping, &config.LocalConfigSampleFileMapping{
		// 	Variable: key,
		// })
	}

	// Save the updated configuration
	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Successfully remapped variables for sample file '%s'.\n", alias)
	return nil
}
