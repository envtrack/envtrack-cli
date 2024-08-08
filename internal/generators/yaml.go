package generators

import (
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type cmdOption struct {
	Name         string
	Shorthand    string `yaml:",omitempty"`
	DefaultValue string `yaml:"default_value,omitempty"`
	Usage        string `yaml:",omitempty"`
}

type CommandDefinition struct {
	Name             string              `yaml:"name"`
	Synopsis         string              `yaml:"synopsis,omitempty"`
	Description      string              `yaml:"description,omitempty"`
	Usage            string              `yaml:"usage,omitempty"`
	Options          []cmdOption         `yaml:"options,omitempty"`
	InheritedOptions []cmdOption         `yaml:"inherited_options,omitempty"`
	Example          string              `yaml:"example,omitempty"`
	SubCommands      []CommandDefinition `yaml:"sub_commands,omitempty"`
}

func GenYamlAllBasic(cmd *cobra.Command, w io.Writer) error {
	return genYAML(cmd, w, true)
}

// GenYamlAll generates a single YAML file containing all commands and subcommands
func GenYamlAll(cmd *cobra.Command, w io.Writer) error {
	return genYAML(cmd, w, false)
}

func genYAML(cmd *cobra.Command, w io.Writer, shorthand bool) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	rootDef := buildCommandDefinition(cmd, shorthand)

	final, err := yaml.Marshal(&rootDef)
	if err != nil {
		return err
	}

	_, err = w.Write(final)
	return err
}

func buildCommandDefinition(cmd *cobra.Command, shorthand bool) CommandDefinition {
	def := CommandDefinition{
		Name: cmd.CommandPath(),
		// Synopsis:    forceMultiLine(cmd.Short),
		Description: forceMultiLine(cmd.Long),
	}
	if !shorthand {
		def.Synopsis = forceMultiLine(cmd.Short)
	}

	if !shorthand && cmd.Runnable() {
		def.Usage = cmd.UseLine()
	}

	if !shorthand && len(cmd.Example) > 0 {
		def.Example = cmd.Example
	}

	flags := cmd.NonInheritedFlags()
	if flags.HasFlags() {
		def.Options = genFlagResult(flags, shorthand)
	}

	flags = cmd.InheritedFlags()
	if flags.HasFlags() {
		def.InheritedOptions = genFlagResult(flags, shorthand)
	}

	for _, subCmd := range cmd.Commands() {
		if !subCmd.IsAvailableCommand() || subCmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		def.SubCommands = append(def.SubCommands, buildCommandDefinition(subCmd, shorthand))
	}

	return def
}

func genFlagResult(flags *pflag.FlagSet, shorthand bool) []cmdOption {
	var result []cmdOption

	flags.VisitAll(func(flag *pflag.Flag) {
		opt := cmdOption{
			Name:         flag.Name,
			DefaultValue: forceMultiLine(flag.DefValue),
			Usage:        forceMultiLine(flag.Usage),
		}
		if !shorthand && len(flag.Shorthand) > 0 && len(flag.ShorthandDeprecated) == 0 {
			opt.Shorthand = flag.Shorthand
		}
		result = append(result, opt)
	})

	return result
}

// GenYamlAllFile generates a single YAML file containing all commands and subcommands
func GenYamlAllFile(cmd *cobra.Command, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return GenYamlAll(cmd, f)
}

// GenYamlAllBasicFile generates a single YAML file containing all commands and subcommands
func GenYamlAllBasicFile(cmd *cobra.Command, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return GenYamlAllBasic(cmd, f)
}

// Temporary workaround for yaml lib generating incorrect yaml with long strings
// that do not contain \n.
func forceMultiLine(s string) string {
	if len(s) > 60 && !strings.Contains(s, "\n") {
		s += "\n"
	}
	return s
}
