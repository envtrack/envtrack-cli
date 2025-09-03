package commandconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/envtrack/envtrack-cli/internal/types"
	"gopkg.in/yaml.v3"
)

// ParsedConfigs represents parsed configuration objects keyed by the original config key.
// Each value is a map[string]interface{} supporting nested maps, arrays and basic types.
type ParsedConfigs map[string]map[string]interface{}

// ParseConfigsFromCommandConfig loads and parses external config files referenced by
// the provided CommandConfig.Configs entries (via CommandConfigConfig.ConfigPath).
// It resolves relative paths against the CommandConfig.RootPath when available,
// otherwise against the current working directory.
func ParseConfigsFromCommandConfig(cmdCfg *types.CommandConfig) (ParsedConfigs, error) {
	parsed := make(ParsedConfigs)
	if cmdCfg == nil || len(cmdCfg.Configs) == 0 {
		return parsed, nil
	}

	for key, cfg := range cmdCfg.Configs {
		if cfg.Location == "" {
			// skip entries without an external config path
			continue
		}

		// Resolve variables (e.g. ${paths.foo}) in the location.
		resolved := resolvePathVariables(cmdCfg, cfg.Location)
		if strings.Contains(cfg.Location, "${") && resolved == "" {
			return parsed, fmt.Errorf("unable to resolve variables in location for key '%s': %s", key, cfg.Location)
		}

		p := resolved
		if p == "" {
			p = cfg.Location
		}

		if !filepath.IsAbs(p) {
			if cmdCfg.RootPath != nil && *cmdCfg.RootPath != "" {
				p = filepath.Join(*cmdCfg.RootPath, p)
			} else {
				cwd, err := os.Getwd()
				if err == nil {
					p = filepath.Join(cwd, p)
				}
			}
		}

		abs, err := filepath.Abs(p)
		if err == nil {
			p = abs
		}

		b, err := os.ReadFile(p)
		if err != nil {
			return parsed, fmt.Errorf("error reading config file for key '%s' at '%s': %w", key, p, err)
		}

		// Try JSON first
		var jm map[string]interface{}
		if err := json.Unmarshal(b, &jm); err == nil {
			parsed[key] = jm
			continue
		}

		// Try YAML
		var yi interface{}
		if err := yaml.Unmarshal(b, &yi); err == nil {
			cv := convertGeneric(yi)
			if m, ok := cv.(map[string]interface{}); ok {
				parsed[key] = m
				continue
			}
			return parsed, fmt.Errorf("parsed YAML for key '%s' is not an object", key)
		}

		return parsed, fmt.Errorf("unable to parse config file for key '%s' as JSON or YAML", key)
	}

	return parsed, nil
}

// ParsedTemplates represents parsed template objects keyed by template name.
// Each value is a map[string]interface{} supporting nested maps, arrays and basic types.
type ParsedTemplates map[string]map[string]interface{}

// ParseTemplatesFromCommandConfig loads and parses template files referenced by
// the provided CommandConfig.CommandTemplatesPaths entries and also includes any
// inlined templates from CommandConfig.Templates. It resolves relative paths
// against the CommandConfig.RootPath when available, otherwise against the
// current working directory.
func ParseTemplatesFromCommandConfig(cmdCfg *types.CommandConfig) (ParsedTemplates, error) {
	parsed := make(ParsedTemplates)
	if cmdCfg == nil {
		return parsed, nil
	}

	// Include any inlined templates first
	if cmdCfg.Templates != nil {
		for name, tmpl := range cmdCfg.Templates {
			// Marshal/unmarshal to convert to map[string]interface{}
			b, err := json.Marshal(tmpl)
			if err != nil {
				// skip templates that cannot be marshalled
				continue
			}
			var m map[string]interface{}
			if err := json.Unmarshal(b, &m); err != nil {
				continue
			}
			parsed[name] = m
		}
	}

	if len(cmdCfg.CommandTemplatesPaths) == 0 {
		return parsed, nil
	}

	for _, origPath := range cmdCfg.CommandTemplatesPaths {
		if origPath == "" {
			continue
		}

		// Resolve variables (e.g. ${paths.foo}) in the path.
		resolved := resolvePathVariables(cmdCfg, origPath)
		if strings.Contains(origPath, "${") && resolved == "" {
			return parsed, fmt.Errorf("unable to resolve variables in template path: %s", origPath)
		}

		p := resolved
		if p == "" {
			p = origPath
		}

		if !filepath.IsAbs(p) {
			if cmdCfg.RootPath != nil && *cmdCfg.RootPath != "" {
				p = filepath.Join(*cmdCfg.RootPath, p)
			} else {
				cwd, err := os.Getwd()
				if err == nil {
					p = filepath.Join(cwd, p)
				}
			}
		}

		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}

		b, err := os.ReadFile(p)
		if err != nil {
			return parsed, fmt.Errorf("error reading template file at '%s': %w", p, err)
		}

		// Try JSON first
		var jm map[string]interface{}
		if err := json.Unmarshal(b, &jm); err == nil {
			// If file contains a top-level 'templates' key, use its contents as templates
			if tval, ok := jm["templates"]; ok {
				cv := convertGeneric(tval)
				if cvM, ok2 := cv.(map[string]interface{}); ok2 {
					for tn, tv := range cvM {
						if tvm, ok3 := tv.(map[string]interface{}); ok3 {
							parsed[tn] = tvm
						} else {
							parsed[tn] = map[string]interface{}{"value": tv}
						}
					}
					continue
				}
			}

			// Otherwise treat top-level object as a mapping of templates
			cv := convertGeneric(jm)
			if cvM, ok := cv.(map[string]interface{}); ok {
				for tn, tv := range cvM {
					if tvm, ok3 := tv.(map[string]interface{}); ok3 {
						parsed[tn] = tvm
					} else {
						parsed[tn] = map[string]interface{}{"value": tv}
					}
				}
				continue
			}
		}

		// Try YAML
		var yi interface{}
		if err := yaml.Unmarshal(b, &yi); err == nil {
			cv := convertGeneric(yi)
			if m, ok := cv.(map[string]interface{}); ok {
				if tval, ok2 := m["templates"]; ok2 {
					if tm, ok3 := tval.(map[string]interface{}); ok3 {
						for tn, tv := range tm {
							if tvm, ok4 := tv.(map[string]interface{}); ok4 {
								parsed[tn] = tvm
							} else {
								parsed[tn] = map[string]interface{}{"value": tv}
							}
						}
						continue
					}
				}
				for tn, tv := range m {
					if tvm, ok3 := tv.(map[string]interface{}); ok3 {
						parsed[tn] = tvm
					} else {
						parsed[tn] = map[string]interface{}{"value": tv}
					}
				}
				continue
			}
		}

		return parsed, fmt.Errorf("unable to parse template file '%s' as JSON or YAML", p)
	}

	return parsed, nil
}

// ParsedCommands represents parsed command objects keyed by command name.
// Each value is a map[string]interface{} supporting nested maps, arrays and basic types.
type ParsedCommands map[string]map[string]interface{}

// ParseCommandsFromCommandConfig loads and parses command definitions referenced by
// the provided CommandConfig.CommandFolderPaths entries and also includes any
// inlined commands from CommandConfig.Commands. It resolves relative paths
// against the CommandConfig.RootPath when available, otherwise against the
// current working directory.
func ParseCommandsFromCommandConfig(cmdCfg *types.CommandConfig) (ParsedCommands, error) {
	parsed := make(ParsedCommands)
	if cmdCfg == nil {
		return parsed, nil
	}

	// Include any inlined commands first
	if cmdCfg.Commands != nil {
		for name, cmd := range cmdCfg.Commands {
			b, err := json.Marshal(cmd)
			if err != nil {
				continue
			}
			var m map[string]interface{}
			if err := json.Unmarshal(b, &m); err != nil {
				continue
			}
			parsed[name] = m
		}
	}

	if len(cmdCfg.CommandFolderPaths) == 0 {
		return parsed, nil
	}

	for _, origPath := range cmdCfg.CommandFolderPaths {
		if origPath == "" {
			continue
		}

		// Resolve variables (e.g. ${paths.foo}) in the path.
		resolved := resolvePathVariables(cmdCfg, origPath)
		if strings.Contains(origPath, "${") && resolved == "" {
			return parsed, fmt.Errorf("unable to resolve variables in command path: %s", origPath)
		}

		p := resolved
		if p == "" {
			p = origPath
		}

		if !filepath.IsAbs(p) {
			if cmdCfg.RootPath != nil && *cmdCfg.RootPath != "" {
				p = filepath.Join(*cmdCfg.RootPath, p)
			} else {
				cwd, err := os.Getwd()
				if err == nil {
					p = filepath.Join(cwd, p)
				}
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
				return parsed, fmt.Errorf("error reading command directory '%s': %v", p, err)
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
				return parsed, fmt.Errorf("error reading command file at '%s': %w", f, err)
			}

			// Try JSON first
			var jm map[string]interface{}
			if err := json.Unmarshal(b, &jm); err == nil {
				// If file contains a top-level 'commands' key, use its contents as commands
				if cval, ok := jm["commands"]; ok {
					cv := convertGeneric(cval)
					if cvM, ok2 := cv.(map[string]interface{}); ok2 {
						for cn, cvv := range cvM {
							if cvm, ok3 := cvv.(map[string]interface{}); ok3 {
								cvm["originalFilepath"] = f
								parsed[cn] = cvm
							} else {
								parsed[cn] = map[string]interface{}{"value": cvv, "originalFilepath": f}
							}
						}
						continue
					}
				}

				// Otherwise treat top-level object as a mapping of commands
				cv := convertGeneric(jm)
				if cvM, ok := cv.(map[string]interface{}); ok {
					for cn, cvv := range cvM {
						if cvm, ok3 := cvv.(map[string]interface{}); ok3 {
							cvm["originalFilepath"] = f
							parsed[cn] = cvm
						} else {
							parsed[cn] = map[string]interface{}{"value": cvv, "originalFilepath": f}
						}
					}
					continue
				}
			}

			// Try YAML
			var yi interface{}
			if err := yaml.Unmarshal(b, &yi); err == nil {
				cv := convertGeneric(yi)
				if m, ok := cv.(map[string]interface{}); ok {
					if cval, ok2 := m["commands"]; ok2 {
						if cm, ok3 := cval.(map[string]interface{}); ok3 {
							for cn, cvv := range cm {
								if cvm, ok4 := cvv.(map[string]interface{}); ok4 {
									cvm["originalFilepath"] = f
									parsed[cn] = cvm
								} else {
									parsed[cn] = map[string]interface{}{"value": cvv, "originalFilepath": f}
								}
							}
							continue
						}
					}
					for cn, cvv := range m {
						if cvm, ok3 := cvv.(map[string]interface{}); ok3 {
							cvm["originalFilepath"] = f
							parsed[cn] = cvm
						} else {
							parsed[cn] = map[string]interface{}{"value": cvv, "originalFilepath": f}
						}
					}
					continue
				}
			}

			return parsed, fmt.Errorf("unable to parse command file '%s' as JSON or YAML", f)
		}
	}

	return parsed, nil
}

// ParseInternalCommandsFromCommandConfig returns a map of types.InternalCommand
// keyed by command name. It mirrors ParseCommandsFromCommandConfig but converts
// the parsed objects into typed InternalCommand instances and sets metadata
// like Name and OriginalFilepath when available.
func ParseInternalCommandsFromCommandConfig(cmdCfg *types.CommandConfig) (map[string]types.InternalCommand, error) {
	result := make(map[string]types.InternalCommand)
	if cmdCfg == nil {
		return result, nil
	}

	// Include inlined commands first
	if cmdCfg.Commands != nil {
		for name, cmd := range cmdCfg.Commands {
			c := cmd // copy
			if c.Name == "" {
				c.Name = name
			}
			// ensure OriginalFilepath points to the source config file when available
			if c.OriginalFilepath == nil && cmdCfg.FilePath != nil {
				c.OriginalFilepath = cmdCfg.FilePath
			}
			result[name] = c
		}
	}

	// Reuse the existing generic parser to gather raw command maps
	parsedGeneric, err := ParseCommandsFromCommandConfig(cmdCfg)
	if err != nil {
		return result, err
	}

	for name, raw := range parsedGeneric {
		// marshal the map back to JSON then unmarshal into the typed struct
		b, merr := json.Marshal(raw)
		if merr != nil {
			// skip this entry if it cannot be marshalled
			continue
		}
		var ic types.InternalCommand
		if uerr := json.Unmarshal(b, &ic); uerr != nil {
			// skip if cannot unmarshal into typed struct
			continue
		}
		if ic.Name == "" {
			ic.Name = name
		}
		// If the raw map included original filepath metadata, prefer that; otherwise leave nil.
		if ofp, ok := raw["originalFilepath"].(string); ok && ofp != "" {
			ic.OriginalFilepath = &ofp
		}

		// If this command references a template, try to locate it and merge (template fields are base, command overrides)
		templateName := ""
		if ic.Template != nil && *ic.Template != "" {
			templateName = *ic.Template
		} else if ic.TemplateName != "" {
			templateName = ic.TemplateName
		} else if rt, ok := raw["template"].(string); ok && rt != "" {
			templateName = rt
		}
		if templateName != "" {
			if tmpl, ok := cmdCfg.Templates[templateName]; ok {
				// merge template and command (override wins)
				merged := mergeInternalCommand(tmpl, ic)
				ic = merged
				// mark that a template was applied
				hasT := true
				ic.HasTemplate = &hasT
			} else {
				// template not found: leave as-is (non-fatal)
			}
		}
		result[name] = ic
	}

	return result, nil
}

// mergeInternalCommand merges a template InternalCommand and an override InternalCommand.
// Fields from override take precedence when set. The template's OriginalTemplateFilepath
// is preserved in the result; OriginalFilepath is taken from the override when present.
func mergeInternalCommand(tmpl types.InternalCommand, override types.InternalCommand) types.InternalCommand {
	res := tmpl // copy

	// Name/FullName
	if override.Name != "" {
		res.Name = override.Name
	}
	if override.FullName != "" {
		res.FullName = override.FullName
	}

	// Pointer booleans and basic pointers: override if non-nil
	if override.SkipTerminalAutofocus != nil {
		res.SkipTerminalAutofocus = override.SkipTerminalAutofocus
	}
	if override.AutocloseCommand != nil {
		res.AutocloseCommand = override.AutocloseCommand
	}
	if override.IsJSONOutput != nil {
		res.IsJSONOutput = override.IsJSONOutput
	}
	if override.HasTemplate != nil {
		res.HasTemplate = override.HasTemplate
	}
	if override.IsInformational != nil {
		res.IsInformational = override.IsInformational
	}
	if override.InformationalCommandSettings != nil {
		res.InformationalCommandSettings = override.InformationalCommandSettings
	}
	if override.KillCmd != nil {
		res.KillCmd = override.KillCmd
	}
	if override.Background != nil {
		res.Background = override.Background
	}
	if override.ForceEnv != nil {
		res.ForceEnv = override.ForceEnv
	}
	if override.Color != nil {
		res.Color = override.Color
	}
	if override.EnableLog != nil {
		res.EnableLog = override.EnableLog
	}
	if override.Description != nil {
		res.Description = override.Description
	}

	// String fields: override if non-empty
	if override.ProtectionOnEnv != "" {
		res.ProtectionOnEnv = override.ProtectionOnEnv
	}
	if override.Command != "" {
		res.Command = override.Command
	}
	if override.Pipe != "" {
		res.Pipe = override.Pipe
	}
	if override.DefaultFileName != "" {
		res.DefaultFileName = override.DefaultFileName
	}
	if override.TemplateName != "" {
		res.TemplateName = override.TemplateName
	}

	// Pointer strings
	if override.Shell != nil {
		res.Shell = override.Shell
	}
	if override.ExternalShellLocation != nil {
		res.ExternalShellLocation = override.ExternalShellLocation
	}
	if override.Template != nil {
		res.Template = override.Template
	}
	if override.Extend != nil {
		res.Extend = override.Extend
	}

	// Slices: fully override when provided (non-nil)
	if override.ExternalShellArguments != nil && len(override.ExternalShellArguments) > 0 {
		res.ExternalShellArguments = override.ExternalShellArguments
	}
	if override.SubCommands != nil && len(override.SubCommands) > 0 {
		res.SubCommands = override.SubCommands
	}
	if override.RequiredParams != nil && len(override.RequiredParams) > 0 {
		res.RequiredParams = override.RequiredParams
	}

	// Params (map[string]string): merge, with override winning
	if res.Params == nil && override.Params == nil {
		// nothing
	} else {
		if res.Params == nil {
			res.Params = make(map[string]string)
		}
		for k, v := range override.Params {
			res.Params[k] = v
		}
	}

	// ParamOptions
	if res.ParamOptions == nil && override.ParamOptions == nil {
		// nothing
	} else {
		if res.ParamOptions == nil {
			res.ParamOptions = make(map[string]types.ParamOption)
		}
		for k, v := range override.ParamOptions {
			res.ParamOptions[k] = v
		}
	}

	// ConfigMap (map[string]string)
	if res.ConfigMap == nil && override.ConfigMap == nil {
		// nothing
	} else {
		if res.ConfigMap == nil {
			res.ConfigMap = make(map[string]string)
		}
		for k, v := range override.ConfigMap {
			res.ConfigMap[k] = v
		}
	}

	// Sessions: override per session key
	if res.Sessions == nil && override.Sessions == nil {
		// nothing
	} else {
		if res.Sessions == nil {
			res.Sessions = make(map[string]types.SessionConfig)
		}
		for k, v := range override.Sessions {
			res.Sessions[k] = v
		}
	}

	// ImportFrom and other interface fields
	if override.ImportFrom != nil {
		res.ImportFrom = override.ImportFrom
	}

	// Original file/template metadata
	// keep template's OriginalTemplateFilepath; override's OriginalFilepath takes precedence
	if tmpl.OriginalTemplateFilepath != nil {
		res.OriginalTemplateFilepath = tmpl.OriginalTemplateFilepath
	}
	if override.OriginalFilepath != nil {
		res.OriginalFilepath = override.OriginalFilepath
	}

	return res
}

// convertGeneric converts YAML-parsed structures (which may contain map[interface{}]interface{})
// into structures using map[string]interface{} recursively so they are JSON-friendly.
func convertGeneric(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{}, len(t))
		for k, val := range t {
			m[k] = convertGeneric(val)
		}
		return m
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(t))
		for k, val := range t {
			m[fmt.Sprint(k)] = convertGeneric(val)
		}
		return m
	case []interface{}:
		arr := make([]interface{}, len(t))
		for i, val := range t {
			arr[i] = convertGeneric(val)
		}
		return arr
	default:
		return t
	}
}

// resolvePathVariables resolves ${...} variables in the provided path string.
// For now it only supports the ${paths.<name>} form and will replace occurrences
// using values from cmdCfg.Paths. If a variable is present but not handled or not
// found, an empty string is returned.
func resolvePathVariables(cmdCfg *types.CommandConfig, input string) string {
	re := regexp.MustCompile(`\$\{([^}]+)}`)
	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return input
	}

	out := input
	for _, m := range matches {
		if len(m) < 2 {
			return ""
		}
		varName := m[1]
		// only support paths.<key> for now
		if strings.HasPrefix(varName, "paths.") {
			key := strings.TrimPrefix(varName, "paths.")
			if cmdCfg != nil && cmdCfg.Paths != nil {
				if val, ok := cmdCfg.Paths[key]; ok {
					out = strings.ReplaceAll(out, m[0], val)
					continue
				}
			}
			// not found
			return ""
		}
		// unsupported variable type
		return ""
	}

	return out
}

// Prevent unused function static analysis warning when cross-package usage isn't detected.
var _ = ParseConfigsFromCommandConfig
var _ = ParseTemplatesFromCommandConfig
var _ = ParseCommandsFromCommandConfig
var _ = ParseInternalCommandsFromCommandConfig
