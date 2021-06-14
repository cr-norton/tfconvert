package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func mergeTemplateFunctions(pfunctions map[string]interface{}) map[string]interface{} {
	functions := map[string]interface{}{
		"formatJSON": formatJSON,
		"lookup":     lookup,
		"tfName":     tfName,
	}
	for k, v := range pfunctions {
		functions[k] = v
	}
	return functions
}

func lookup(stack Stack, id string) string {
	if res := stack.Lookup(id); res != nil {
		return fmt.Sprintf("%s.%s.%s", res.Type, tfName(res.Identifier), res.OutputKey)
	}
	return fmt.Sprintf(`"%s"`, id)
}

func formatJSON(str string) string {
	str = strings.ReplaceAll(str, `\`, ``)
	var out bytes.Buffer
	json.Indent(&out, []byte(str), "", "  ")
	return out.String()
}

func tfName(name string) string {
	return toSnakeCase(name)
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
