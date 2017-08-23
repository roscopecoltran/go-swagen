package typescript

import "github.com/go-openapi/spec"
import "github.com/xreception/go-swagen/utils"

func parameterType(param *spec.Parameter) string {
	if param.Schema != nil {
		return schemaType(param.Schema)
	}

	if param.Type == "integer" {
		return "number"
	}

	if param.Type == "array" {
		return "Array<" + param.Items.Type + ">"
	}

	return param.Type
}

func responseType(resp spec.Response) string {
	return schemaType(resp.Schema)
}

func schemaRef(schema spec.Schema) *spec.Schema {
	return &schema
}

func schemaType(schema *spec.Schema) string {
	if schema == nil {
		return ""
	}

	if schema.Ref.HasFragmentOnly {
		return utils.InterfaceCase(schemaName(schema))
	}

	if len(schema.Type) == 0 {
		return ""
	}

	if schema.Type[0] == "integer" {
		return "number"
	}

	if schema.Type[0] == "array" {
		return schemaType(schema.Items.Schema) + "[]"
	}

	return schema.Type[0]
}

func schemaName(schema *spec.Schema) string {
	pr := schema.Ref.GetPointer()
	tokens := pr.DecodedTokens()
	return tokens[len(tokens)-1]
}
