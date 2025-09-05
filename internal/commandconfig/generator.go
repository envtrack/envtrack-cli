package commandconfig

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/envtrack/envtrack-cli/internal/types"
	"github.com/envtrack/envtrack-cli/internal/variable_parser"
	"github.com/hoisie/mustache"
)

// BuildCommandString constructs the final command string for the provided
// InternalCommand by merging default params, session params and supplied params
// (supplied params take precedence). It renders mustache-style templates
// (e.g. {{params.foo}}) and also replaces ${...} variables (e.g. ${params.foo} or ${foo}).
// For bare ${name} references, merged params (from ResolveParams) are checked first.
// The full command configuration is accepted as cmdCfg to allow resolving references
// into parsed configs (configs.*). Returns the rendered command string, any warning
// messages produced while resolving params, and an error for fatal unresolved variables.
func BuildCommandString(cmdCfg *types.CommandConfig, ic types.InternalCommand, sessionName string, suppliedParams map[string]string) (string, []string, error) {
	warnings := []string{}

	// Resolve merged params using helper (currently reads from ic.Params)
	mergedParams := ResolveParams(ic, sessionName, suppliedParams)

	// Parse external config files referenced by the command config so we can resolve configs.* refs
	parsedCfgs, perr := ParseConfigsFromCommandConfig(cmdCfg)
	if perr != nil {
		// non-fatal: include warning and continue with whatever parsedCfgs were returned
		warnings = append(warnings, fmt.Sprintf("warning parsing referenced configs: %v", perr))
	}

	// Resolve parameter references like configs.ENV.values.FOO
	resolvedParams, refsWarnings := ResolveParamReferences(cmdCfg, ic, sessionName, mergedParams, parsedCfgs)
	if len(refsWarnings) > 0 {
		warnings = append(warnings, refsWarnings...)
	}

	// Merge config maps similarly (session overrides command)
	mergedConfig := make(map[string]string)
	for k, v := range ic.ConfigMap {
		mergedConfig[k] = v
	}
	if sessionName != "" && ic.Sessions != nil {
		if sc, ok := ic.Sessions[sessionName]; ok {
			for k, v := range sc.ConfigMap {
				mergedConfig[k] = v
			}
		}
	}

	// Prepare data for variable resolution
	data := map[string]interface{}{
		"params": resolvedParams,
		"config": mergedConfig,
		"session": map[string]interface{}{
			"name": sessionName,
		},
	}

	// Include paths from cmdCfg (if present) so callers can reference ${paths.xxx}
	if cmdCfg != nil && cmdCfg.Paths != nil {
		data["paths"] = cmdCfg.Paths
	}

	vm := variableparser.NewVariableMapper(data, nil)

	// mustache render: provide a function that returns string values for variables
	renderFunc := func(name string) string {
		if v, ok := vm.ResolveVariable(name); ok {
			return fmt.Sprint(v)
		}
		return ""
	}

	// First render mustache-style templates
	rendered := mustache.Render(ic.Command, renderFunc)

	// Then replace ${...} style variables
	re := regexp.MustCompile(`\$\{([^}]+)}`)
	unresolved := make([]string, 0)
	replaced := re.ReplaceAllStringFunc(rendered, func(m string) string {
		sub := strings.TrimSuffix(strings.TrimPrefix(m, "${"), "}")
		// Check merged/resolved params for bare names (e.g. ${foo}) first
		if v, ok := resolvedParams[sub]; ok {
			return v
		}
		// Fallback: allow full variable resolution (e.g. params.foo, config.bar, paths.xxx)
		if v, ok := vm.ResolveVariable(sub); ok {
			return fmt.Sprint(v)
		}
		unresolved = append(unresolved, sub)
		return ""
	})

	if len(unresolved) > 0 {
		return replaced, warnings, fmt.Errorf("unresolved variables: %v", unresolved)
	}

	return replaced, warnings, nil
}
