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

		// Load templates from any CommandTemplatesPaths into baseCfg.Templates
		if err := loadTemplatesIntoCommandConfig(baseCfg); err != nil {
			// Non-fatal: log and continue returning the merged config
			log.Printf("Warning: error loading templates: %v\n", err)
		}

		// Load commands from any CommandFolderPaths into baseCfg.Commands
		if err := loadCommandsIntoCommandConfig(baseCfg); err != nil {
			log.Printf("Warning: error loading commands: %v\n", err)
		}

		return baseCfg, nil
	} else {
		log.Println("No external config path set; using parsed config as-is.")
	}

	// Load templates from any CommandTemplatesPaths into the parsed commandConfig
	if err := loadTemplatesIntoCommandConfig(&commandConfig); err != nil {
		log.Printf("Warning: error loading templates: %v\n", err)
	}

	// Load commands from any CommandFolderPaths into the parsed commandConfig
	if err := loadCommandsIntoCommandConfig(&commandConfig); err != nil {
		log.Printf("Warning: error loading commands: %v\n", err)
	}

	return &commandConfig, nil
}

// loadTemplatesIntoCommandConfig scans CommandTemplatesPaths (files or directories)
// and loads template definitions into cfg.Templates. YAML/JSON files are supported.
// Non-fatal errors are returned but callers typically log and continue.
func loadTemplatesIntoCommandConfig(cfg *types.CommandConfig) error {
	if cfg == nil || len(cfg.CommandTemplatesPaths) == 0 {
		return nil
	}

	if cfg.Templates == nil {
		cfg.Templates = make(map[string]types.InternalCommand)
	}

	var aggregateErrs []string

	for _, pathEntry := range cfg.CommandTemplatesPaths {
		if pathEntry == "" {
			continue
		}

		// Resolve variables in pathEntry
		resolved := resolvePathVariables(cfg, pathEntry)
		if strings.Contains(pathEntry, "${") && resolved == "" {
			aggregateErrs = append(aggregateErrs, fmt.Sprintf("unable to resolve variables in template path: %s", pathEntry))
			continue
		}
		p := resolved
		if p == "" {
			p = pathEntry
		}

		if !filepath.IsAbs(p) {
			if cfg.RootPath != nil && *cfg.RootPath != "" {
				p = filepath.Join(*cfg.RootPath, p)
			} else if cwd, err := os.Getwd(); err == nil {
				p = filepath.Join(cwd, p)
			}
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}

		fi, err := os.Stat(p)
		if err != nil {
			//aggregateErrs = append(aggregateErrs, fmt.Sprintf("error accessing template path '%s': %v", p, err))
			continue
		}

		var files []string
		if fi.IsDir() {
			entries, err := os.ReadDir(p)
			if err != nil {
				aggregateErrs = append(aggregateErrs, fmt.Sprintf("error reading template directory '%s': %v", p, err))
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(e.Name()))
				if ext == ".yaml" || ext == ".yml" || ext == ".json" {
					files = append(files, filepath.Join(p, e.Name()))
				}
			}
		} else {
			files = append(files, p)
		}

		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				aggregateErrs = append(aggregateErrs, fmt.Sprintf("error reading template file '%s': %v", f, err))
				continue
			}

			// Try YAML/JSON unmarshal into either {templates: {...}} or a top-level map of templates
			// First try wrapper { templates: map[string]InternalCommand }
			var wrapper struct {
				Templates map[string]types.InternalCommand `json:"templates" yaml:"templates"`
			}
			if err := yaml.Unmarshal(b, &wrapper); err == nil && len(wrapper.Templates) > 0 {
				for tn, tc := range wrapper.Templates {
					// set metadata
					tc.Name = tn
					tc.OriginalTemplateFilepath = &f
					cfg.Templates[tn] = tc
				}
				continue
			}

			// Try direct mapping
			var direct map[string]types.InternalCommand
			if err := yaml.Unmarshal(b, &direct); err == nil && len(direct) > 0 {
				for tn, tc := range direct {
					tc.Name = tn
					tc.OriginalTemplateFilepath = &f
					cfg.Templates[tn] = tc
				}
				continue
			}

			// As a fallback try JSON unmarshal (for .json files or other cases)
			if err := json.Unmarshal(b, &wrapper); err == nil && len(wrapper.Templates) > 0 {
				for tn, tc := range wrapper.Templates {
					tc.Name = tn
					tc.OriginalTemplateFilepath = &f
					cfg.Templates[tn] = tc
				}
				continue
			}
			if err := json.Unmarshal(b, &direct); err == nil && len(direct) > 0 {
				for tn, tc := range direct {
					tc.Name = tn
					tc.OriginalTemplateFilepath = &f
					cfg.Templates[tn] = tc
				}
				continue
			}

			aggregateErrs = append(aggregateErrs, fmt.Sprintf("unable to parse template file '%s' as JSON or YAML", f))
		}
	}

	if len(aggregateErrs) > 0 {
		return fmt.Errorf(strings.Join(aggregateErrs, "; "))
	}
	return nil
}

// loadCommandsIntoCommandConfig scans CommandFolderPaths (files or directories)
// and loads command definitions into cfg.Commands. YAML/JSON files are supported.
// Non-fatal errors are returned but callers typically log and continue.
func loadCommandsIntoCommandConfig(cfg *types.CommandConfig) error {
	if cfg == nil || len(cfg.CommandFolderPaths) == 0 {
		return nil
	}

	if cfg.Commands == nil {
		cfg.Commands = make(map[string]types.InternalCommand)
	}

	var aggregateErrs []string

	for _, pathEntry := range cfg.CommandFolderPaths {
		if pathEntry == "" {
			continue
		}

		// Resolve variables in pathEntry
		resolved := resolvePathVariables(cfg, pathEntry)
		if strings.Contains(pathEntry, "${") && resolved == "" {
			aggregateErrs = append(aggregateErrs, fmt.Sprintf("unable to resolve variables in command path: %s", pathEntry))
			continue
		}
		p := resolved
		if p == "" {
			p = pathEntry
		}

		if !filepath.IsAbs(p) {
			if cfg.RootPath != nil && *cfg.RootPath != "" {
				p = filepath.Join(*cfg.RootPath, p)
			} else if cwd, err := os.Getwd(); err == nil {
				p = filepath.Join(cwd, p)
			}
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}

		fi, err := os.Stat(p)
		if err != nil {
			// skip non-existent paths
			continue
		}

		var files []string
		if fi.IsDir() {
			entries, err := os.ReadDir(p)
			if err != nil {
				aggregateErrs = append(aggregateErrs, fmt.Sprintf("error reading command directory '%s': %v", p, err))
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				ext := strings.ToLower(filepath.Ext(e.Name()))
				if ext == ".yaml" || ext == ".yml" || ext == ".json" {
					files = append(files, filepath.Join(p, e.Name()))
				}
			}
		} else {
			files = append(files, p)
		}

		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				aggregateErrs = append(aggregateErrs, fmt.Sprintf("error reading command file '%s': %v", f, err))
				continue
			}

			// Try YAML/JSON unmarshal into either {commands: {...}} or a top-level map of commands
			// First try wrapper { commands: map[string]InternalCommand }
			var wrapper struct {
				Commands map[string]types.InternalCommand `json:"commands" yaml:"commands"`
			}
			if err := yaml.Unmarshal(b, &wrapper); err == nil && len(wrapper.Commands) > 0 {
				for cn, cc := range wrapper.Commands {
					cc.Name = cn
					cc.OriginalFilepath = &f
					cfg.Commands[cn] = cc
				}
				continue
			}

			// Try direct mapping
			var direct map[string]types.InternalCommand
			if err := yaml.Unmarshal(b, &direct); err == nil && len(direct) > 0 {
				for cn, cc := range direct {
					cc.Name = cn
					cc.OriginalFilepath = &f
					cfg.Commands[cn] = cc
				}
				continue
			}

			// As a fallback try JSON unmarshal (for .json files or other cases)
			if err := json.Unmarshal(b, &wrapper); err == nil && len(wrapper.Commands) > 0 {
				for cn, cc := range wrapper.Commands {
					cc.Name = cn
					cc.OriginalFilepath = &f
					cfg.Commands[cn] = cc
				}
				continue
			}
			if err := json.Unmarshal(b, &direct); err == nil && len(direct) > 0 {
				for cn, cc := range direct {
					cc.Name = cn
					cc.OriginalFilepath = &f
					cfg.Commands[cn] = cc
				}
				continue
			}

			aggregateErrs = append(aggregateErrs, fmt.Sprintf("unable to parse command file '%s' as JSON or YAML", f))
		}
	}

	if len(aggregateErrs) > 0 {
		return fmt.Errorf(strings.Join(aggregateErrs, "; "))
	}
	return nil
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
	if overlay.CommandFolderPaths != nil {
		if base.CommandFolderPaths == nil {
			base.CommandFolderPaths = make([]string, 0)
		}
		base.CommandFolderPaths = append(base.CommandFolderPaths, overlay.CommandFolderPaths...)
	}
	if overlay.Commands != nil {
		if base.Commands == nil {
			base.Commands = make(map[string]types.InternalCommand)
		}
		for k, v := range overlay.Commands {
			base.Commands[k] = v
		}
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
