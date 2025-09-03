package commandconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/envtrack/envtrack-cli/internal/types"
	"gopkg.in/yaml.v3"
)

// LoadFromFile loads and parses a command configuration file (JSON or YAML).
// It returns a fully parsed types.CommandConfig or an error.
func LoadFromFile(path string) (*types.CommandConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("Error closing config file: %v\n", cerr)
		}
	}()

	var commandConfig types.CommandConfig

	ext := filepath.Ext(path)
	if ext == ".yaml" || ext == ".yml" {
		decoder := yaml.NewDecoder(file)
		if err := decoder.Decode(&commandConfig); err != nil {
			return nil, fmt.Errorf("error parsing YAML config file: %w", err)
		}
	} else {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&commandConfig); err != nil {
			return nil, fmt.Errorf("error parsing JSON config file: %w", err)
		}
	}

	abs, err := filepath.Abs(path)
	if err == nil {
		commandConfig.FilePath = &abs
		root := filepath.Dir(abs)
		commandConfig.RootPath = &root
	} else {
		p := path
		commandConfig.FilePath = &p
	}

	// If an externalConfigPath is set on the parsed config, load that first as the base
	// and then merge the values from the current (parsed) config into it (overlay wins).
	if commandConfig.ExternalConfigPath != nil && *commandConfig.ExternalConfigPath != "" {
		log.Printf("Using external config path: %s\n", *commandConfig.ExternalConfigPath)
		// Resolve external path relative to the current config's root path when possible
		external := *commandConfig.ExternalConfigPath

		// Try to resolve variable references like ${paths.xxx} using commandConfig.Paths
		if resolved := resolvePathVariables(&commandConfig, external); resolved != "" {
			external = resolved
		} else if strings.Contains(external, "${") {
			// variable present but couldn't resolve
			return nil, fmt.Errorf("unable to resolve variables in externalConfigPath: %s", external)
		}
		if !filepath.IsAbs(external) {
			if commandConfig.RootPath != nil && *commandConfig.RootPath != "" {
				external = filepath.Join(*commandConfig.RootPath, external)
			} else {
				if cwd, cerr := os.Getwd(); cerr == nil {
					external = filepath.Join(cwd, external)
				}
			}
		}

		if eabs, err := filepath.Abs(external); err == nil {
			external = eabs
		}

		baseCfg, berr := LoadFromFile(external)
		if berr != nil {
			return nil, fmt.Errorf("error loading external config '%s': %w", external, berr)
		}

		// Merge parsed (overlay) into baseCfg
		mergeCommandConfig(baseCfg, &commandConfig)

		// Ensure resulting config has FilePath/RootPath from the original parsed config
		baseCfg.FilePath = commandConfig.FilePath
		baseCfg.RootPath = commandConfig.RootPath

		return baseCfg, nil
	} else {
		log.Println("No external config path set; using parsed config as-is.")
	}

	return &commandConfig, nil
}

// mergeCommandConfig merges overlay into base. overlay values overwrite base where applicable.
// For slices, values are appended. For maps, keys from overlay override existing keys; nested
// maps (e.g., Variables) are merged with overlay values overwriting matching keys.
func mergeCommandConfig(base, overlay *types.CommandConfig) {
	if overlay == nil || base == nil {
		return
	}

	// Simple pointer fields: prefer overlay if set
	if overlay.ExternalConfigPath != nil {
		base.ExternalConfigPath = overlay.ExternalConfigPath
	}
	if overlay.FilePath != nil {
		base.FilePath = overlay.FilePath
	}
	if overlay.RootPath != nil {
		base.RootPath = overlay.RootPath
	}
	if overlay.ProjectRootPath != nil {
		base.ProjectRootPath = overlay.ProjectRootPath
	}
	if overlay.ConfigPath != nil {
		base.ConfigPath = overlay.ConfigPath
	}
	if overlay.DefaultSession != nil {
		base.DefaultSession = overlay.DefaultSession
	}
	if overlay.AutoSelectDefaultSession != nil {
		base.AutoSelectDefaultSession = overlay.AutoSelectDefaultSession
	}

	// Slices: append
	base.CommandTemplatesPaths = append(base.CommandTemplatesPaths, overlay.CommandTemplatesPaths...)
	base.CommandFolderPaths = append(base.CommandFolderPaths, overlay.CommandFolderPaths...)
	base.ObservableCommands = append(base.ObservableCommands, overlay.ObservableCommands...)

	// Maps: merge/overwrite
	if base.Commands == nil && overlay.Commands != nil {
		base.Commands = make(map[string]types.InternalCommand)
	}
	for k, v := range overlay.Commands {
		base.Commands[k] = v
	}

	if base.Templates == nil && overlay.Templates != nil {
		base.Templates = make(map[string]types.InternalCommand)
	}
	for k, v := range overlay.Templates {
		base.Templates[k] = v
	}

	if base.Autostart == nil && overlay.Autostart != nil {
		base.Autostart = make(map[string]*string)
	}
	for k, v := range overlay.Autostart {
		base.Autostart[k] = v
	}

	if base.Paths == nil && overlay.Paths != nil {
		base.Paths = make(map[string]string)
	}
	for k, v := range overlay.Paths {
		base.Paths[k] = v
	}

	// Configs: merge per-key; for Variables map, merge keys with overlay overwriting
	if base.Configs == nil && overlay.Configs != nil {
		base.Configs = make(map[string]types.CommandConfigConfig)
	}
	for k, ov := range overlay.Configs {
		if bv, ok := base.Configs[k]; ok {
			// merge
			if ov.Location != "" {
				bv.Location = ov.Location
			}
			if bv.Variables == nil && ov.Variables != nil {
				bv.Variables = make(map[string]interface{})
			}
			for vk, vv := range ov.Variables {
				bv.Variables[vk] = vv
			}
			if ov.ConfigPath != nil {
				bv.ConfigPath = ov.ConfigPath
			}
			base.Configs[k] = bv
		} else {
			base.Configs[k] = ov
		}
	}

	// ObservableCommands already appended above
}

// ValidateConfigFile attempts to load and parse the config file to ensure it is valid.
func ValidateConfigFile(path string) error {
	_, err := LoadFromFile(path)
	return err
}

// Prevent unused function static analysis warning when cross-package usage isn't detected.
var _ = ValidateConfigFile
