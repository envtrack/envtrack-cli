package variableparser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hoisie/mustache"
)

type VariableMapper struct {
	Data     map[string]interface{}
	Mappings map[string]string
}

func NewVariableMapper(data map[string]interface{}, mappings map[string]string) *VariableMapper {
	return &VariableMapper{
		Data:     data,
		Mappings: mappings,
	}
}

func (vm *VariableMapper) ResolveVariable(name string) (interface{}, bool) {
	// Check if there's a mapping for this variable
	if mappedName, ok := vm.Mappings[name]; ok {
		name = mappedName
	}

	parts := strings.Split(name, ".")
	current := reflect.ValueOf(vm.Data)

	for _, part := range parts {
		if current.Kind() == reflect.Map {
			current = current.MapIndex(reflect.ValueOf(part))
		} else if current.Kind() == reflect.Struct {
			current = current.FieldByName(part)
		} else {
			return nil, false
		}

		if !current.IsValid() {
			return nil, false
		}
	}

	return current.Interface(), true
}

func (vm *VariableMapper) ValidateAndOutputVariables(template string) ([]string, error) {
	var variables []string
	var invalidVariables []string

	// Custom function to capture variable names
	captureVar := func(name string) string {
		if _, ok := vm.ResolveVariable(name); !ok {
			invalidVariables = append(invalidVariables, name)
		}
		variables = append(variables, name)
		return ""
	}

	// Use mustache.Render with our custom function
	mustache.Render(template, map[string]interface{}{
		"": captureVar,
	})

	if len(invalidVariables) > 0 {
		return variables, fmt.Errorf("invalid variables: %v", invalidVariables)
	}

	return variables, nil
}

// func main() {
// 	data := map[string]interface{}{
// 		"env": map[string]string{
// 			"var1": "value1",
// 			"var2": "value2",
// 		},
// 		"config": map[string]string{
// 			"setting1": "settingValue1",
// 		},
// 	}

// 	vm := NewVariableMapper(data)

// 	// Add mappings
// 	vm.AddMapping("foo.bar", "env.var1")
// 	vm.AddMapping("foo", "env")

// 	template := "Hello {{foo.bar}} and {{config.setting1}} and {{env.var2}}"

// 	variables, err := vm.ValidateAndOutputVariables(template)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 	} else {
// 		fmt.Println("Valid variables:", variables)
// 	}

// 	// Render the template
// 	result := mustache.Render(template, vm.ResolveVariable)
// 	fmt.Println("Rendered result:", result)
// }
