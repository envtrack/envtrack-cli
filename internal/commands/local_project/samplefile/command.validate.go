package samplefile

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func validateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <alias>",
		Args:  cobra.ExactArgs(1),
		Short: "Validate a sample file",
		RunE:  runRemap,
	}
	return cmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	if len(args) != 1 {
		return fmt.Errorf("invalid number of arguments")
	}
	alias := args[0]

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
