package commandconfig

import (
	"fmt"
	"strings"

	"github.com/envtrack/envtrack-cli/internal/types"
)

// ResolveParams returns merged parameters for the given InternalCommand.
// Merge order: command default params -> session params -> supplied params (supplied overrides).
func ResolveParams(ic types.InternalCommand, sessionName string, supplied map[string]string) map[string]string {
	merged := make(map[string]string)
	// command-level defaults
	for k, v := range ic.Params {
		merged[k] = v
	}
	// session overrides
	if sessionName != "" && ic.Sessions != nil {
		if sc, ok := ic.Sessions[sessionName]; ok {
			for k, v := range sc.Params {
				merged[k] = v
			}
		}
	}
	// supplied overrides
	for k, v := range supplied {
		merged[k] = v
	}
	return merged
}

// ResolveParamReferences inspects parameter values and resolves references of the
// form "configs.<configMapKey>.<path.to.value>" using the provided parsedConfigs
// (from ParseConfigsFromCommandConfig) and the command's ConfigMap mapping.
// It returns a possibly-updated params map and a slice of warning messages for
// lookups that could not be resolved.
func ResolveParamReferences(cmdCfg *types.CommandConfig, ic types.InternalCommand, sessionName string, params map[string]string, parsedConfigs ParsedConfigs) (map[string]string, []string) {
	warnings := []string{}
	out := make(map[string]string)
	for k, v := range params {
		// only handle values that start with "configs." explicitly
		if !strings.HasPrefix(v, "configs.") {
			out[k] = v
			continue
		}

		parts := strings.Split(v, ".")
		if len(parts) < 2 {
			warnings = append(warnings, fmt.Sprintf("param '%s' has invalid configs reference '%s'", k, v))
			out[k] = v
			continue
		}
		// parts[0] == "configs" by prefix check
		configMapKey := parts[1] // e.g. "ENV"
		var mappedCfgName string
		var ok bool

		// Prefer session-specific configMap mapping when a session is provided
		if sessionName != "" && ic.Sessions != nil {
			if sc, sok := ic.Sessions[sessionName]; sok {
				mappedCfgName, ok = sc.ConfigMap[configMapKey]
				if ok && mappedCfgName != "" {
					// found in session configMap
				} else {
					// not found in session; fallthrough to command-level mapping
					ok = false
				}
			}
		}

		// If not found in session, check command-level configMap
		if !ok {
			mappedCfgName, ok = ic.ConfigMap[configMapKey]
		}

		// Fallback: if not present on the command, check top-level cmdCfg.Configs for a key with that name
		if (!ok || mappedCfgName == "") && cmdCfg != nil {
			if _, exists := cmdCfg.Configs[configMapKey]; exists {
				mappedCfgName = configMapKey
				ok = true
			}
		}

		if !ok || mappedCfgName == "" {
			warnings = append(warnings, fmt.Sprintf("param '%s' references configs.%s but no mapping found in session or command configMap or top-level configs", k, configMapKey))
			out[k] = v
			continue
		}
		// Look up the parsed config by the mapped name
		pcfg, ok := parsedConfigs[mappedCfgName]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("param '%s' references configs.%s mapped to '%s', but no parsed config found for '%s'", k, configMapKey, mappedCfgName, mappedCfgName))
			out[k] = v
			continue
		}
		// Traverse remaining path segments (parts[2:], e.g. values, ORB_DB_USERNAME)
		if len(parts) < 3 {
			warnings = append(warnings, fmt.Sprintf("param '%s' references configs.%s but no path provided: '%s'", k, configMapKey, v))
			out[k] = v
			continue
		}
		curr := interface{}(pcfg)
		pathParts := parts[2:]
		resolved := true
		for _, p := range pathParts {
			switch cm := curr.(type) {
			case map[string]interface{}:
				if nv, ok := cm[p]; ok {
					curr = nv
				} else {
					resolved = false
					break
				}
			default:
				resolved = false
				break
			}
			if !resolved {
				break
			}
		}
		if !resolved {
			warnings = append(warnings, fmt.Sprintf("param '%s' path '%s' not found in parsed config '%s'", k, strings.Join(parts[2:], "."), mappedCfgName))
			out[k] = v
			continue
		}
		// success: convert curr to string
		out[k] = fmt.Sprint(curr)
	}
	return out, warnings
}
