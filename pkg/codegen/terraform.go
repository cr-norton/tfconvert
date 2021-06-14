package codegen

import (
	"fmt"

	"github.com/cr-norton/tfconvert/pkg/types"
)

func GenerateImportScript(resources []types.Resource) ([]string, error) {
	commands := []string{}
	for _, resource := range resources {
		name := tfName(resource.Identifier)
		command := fmt.Sprintf("terraform import %s.%s %s", resource.Type, name, resource.ImportKey)
		commands = append(commands, command)
	}
	return commands, nil
}
