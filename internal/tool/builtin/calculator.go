package builtin

import (
	"MyAgent/internal/tool"
	"context"
	"encoding/json"
	"fmt"
)

var _ tool.Tool = (*CalculatorTool)(nil)

const calculatorSchema = `{
  "type": "object",
  "properties": {
    "a": {
      "type": "number",
      "description": "第一个数字"
    },
    "b": {
      "type": "number",
      "description": "第二个数字"
    },
    "operation": {
      "type": "string",
      "enum": ["add", "subtract", "multiply"],
      "description": "运算类型"
    }
  },
  "required": ["a", "b", "operation"],
  "additionalProperties": false
}`

type CalculatorTool struct{}

func (c *CalculatorTool) Definition() tool.Definition {
	return tool.Definition{
		Name:        "calculator",
		Description: "执行基本数学运算",
		Parameters:  json.RawMessage(calculatorSchema),
	}
}

func (c *CalculatorTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args calculatorArgs

	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	var result float64

	switch args.Operation {
	case "add":
		result = args.A + args.B
	case "subtract":
		result = args.A - args.B
	case "multiply":
		result = args.A * args.B
	default:
		return "", fmt.Errorf(
			"unsupported operation: %s",
			args.Operation,
		)
	}

	return fmt.Sprintf("%v", result), nil
}

type calculatorArgs struct {
	A         float64 `json:"a"`
	B         float64 `json:"b"`
	Operation string  `json:"operation"`
}
