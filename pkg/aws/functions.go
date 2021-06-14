package aws

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var TemplateFunctions = map[string]interface{}{
	"keySchemaElement":    keySchemaElement,
	"parsePolicyDocument": parsePolicyDocument,
	"typeStringSlice":     typeStringSlice,
}

func keySchemaElement(schema []types.KeySchemaElement, ktype types.KeyType) *string {
	for _, kschema := range schema {
		if kschema.KeyType == ktype {
			return kschema.AttributeName
		}
	}
	return nil
}

// ! being lazy here since templates aren't typesafe anyway
func parsePolicyDocument(policyDocument string) interface{} {
	decoded, err := url.QueryUnescape(policyDocument)
	if err != nil {
		panic(err) // TODO
	}

	// TODO
	var output interface{}
	json.Unmarshal([]byte(decoded), &output)
	return output
}

// TODO move
func typeStringSlice(v interface{}) []string {
	if str, ok := v.(string); ok {
		str = strings.ReplaceAll(str, "[", "")
		str = strings.ReplaceAll(str, "]", "")
		return strings.Split(str, " ")
	}

	if vals, ok := v.([]interface{}); ok {
		strs := []string{}
		for _, val := range vals {
			if sval, ok := val.(string); ok {
				strs = append(strs, sval)
			}
		}
		return strs
	}

	return []string{}
}
