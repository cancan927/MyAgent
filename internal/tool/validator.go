package tool

import (
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

type Validator interface {
	Validate(schema []byte, data []byte) error
}

type JSONSchemaValidator struct{}

func (v *JSONSchemaValidator) Validate(schema []byte, data []byte) error {
	schemaLoader := gojsonschema.NewBytesLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validate schema: %w", err)
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			return fmt.Errorf(
				"invalid tool arguments: %s",
				desc.String(),
			)
		}
	}

	return nil
}
