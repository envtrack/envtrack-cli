package samplefile

import (
	"fmt"
	"os"
	"sort"

	"github.com/cbroglie/mustache"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func addSampleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Links a local sample environment file to the local configuration",
		RunE:  runAddSample,
	}
	cmd.Flags().String("file", "", "Path to the sample environment file (required)")
	cmd.Flags().StringP("name", "n", "", "Name for the sample file in the configuration (required)")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("name")
	return cmd
}

func runAddSample(cmd *cobra.Command, args []string) error {
	localCfg, err := config.LocalConf.GetLocalConfig()
	if err != nil {
		return fmt.Errorf("no local context found. Use 'envtrack init' to initialize a local project")
	}

	filePath, _ := cmd.Flags().GetString("file")
	name, _ := cmd.Flags().GetString("name")

	env, err := localCfg.GetSelectedEnvironment()
	if err != nil {
		return fmt.Errorf("no environment selected. Use 'envtrack use' to select an environment")
	}

	// Check if the sample name already exists
	for _, sample := range env.SampleFiles {
		if sample.Alias == name {
			return fmt.Errorf("sample file with name '%s' already exists", name)
		}
	}

	// Create the new sample file entry
	newSample := &config.LocalConfigSampleFile{
		Alias:     name,
		Path:      filePath,
		Variables: []string{},
	}

	varMapping, err := createMapping(filePath)
	if err != nil {
		return err
	}
	for _, key := range varMapping {
		newSample.Variables = append(newSample.Variables, key)
		// newSample.Mapping = append(newSample.Mapping, &config.LocalConfigSampleFileMapping{
		// 	Variable: key,
		// })
	}

	// Add the new sample to the configuration
	env.SampleFiles = append(env.SampleFiles, newSample)

	// Save the updated configuration
	err = config.LocalConf.SaveLocalConfig(*localCfg)
	if err != nil {
		return fmt.Errorf("error saving local configuration: %v", err)
	}

	fmt.Printf("Successfully added sample file '%s' with name '%s'.\n", filePath, name)
	return nil
}

func createMapping(filePath string) ([]string, error) {
	// Read the entire file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse the content as a mustache template
	template, err := mustache.ParseString(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing mustache template: %v", err)
	}

	// Get all tags from the template
	tags := template.Tags()

	// Convert tags to a slice and sort them
	variables := make([]string, 0, len(tags))
	for _, tag := range tags {
		variables = append(variables, tag.Name())
	}
	sort.Strings(variables)

	return variables, nil
}
