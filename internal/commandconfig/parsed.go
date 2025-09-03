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
