package shape_test

import (
	"fmt"

	"github.com/shapestone/shape/internal/parser"
	"github.com/shapestone/shape/pkg/ast"
	"github.com/shapestone/shape/pkg/shape"
)

func ExampleParse() {
	input := `{"id": UUID, "name": String(1, 100)}`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		panic(err)
	}

	fmt.Println(node.Type())
	// Output: Object
}

func ExampleParseAuto() {
	input := `{"id": UUID, "email": Email}`

	node, format, err := shape.ParseAuto(input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Format: %s, Type: %s\n", format, node.Type())
	// Output: Format: JSONV, Type: Object
}

func ExampleMustParse() {
	// MustParse is useful in tests or initialization where input is known to be valid
	node := shape.MustParse(parser.FormatJSONV, `{"id": UUID}`)

	obj := node.(*ast.ObjectNode)
	fmt.Println(len(obj.Properties()))
	// Output: 1
}

func ExampleParse_nestedObject() {
	input := `{
		"user": {
			"id": UUID,
			"profile": {
				"name": String(1, 100),
				"email": Email
			}
		}
	}`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		panic(err)
	}

	root := node.(*ast.ObjectNode)
	user, _ := root.GetProperty("user")
	userObj := user.(*ast.ObjectNode)
	profile, _ := userObj.GetProperty("profile")

	fmt.Println(profile.Type())
	// Output: Object
}

func ExampleParse_array() {
	input := `[String(1, 30)]`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		panic(err)
	}

	arr := node.(*ast.ArrayNode)
	elem := arr.ElementSchema()
	fn := elem.(*ast.FunctionNode)

	fmt.Printf("Array element: %s\n", fn.Name())
	// Output: Array element: String
}

func ExampleParse_function() {
	input := `Integer(18, 120)`

	node, err := shape.Parse(parser.FormatJSONV, input)
	if err != nil {
		panic(err)
	}

	fn := node.(*ast.FunctionNode)
	fmt.Printf("Function: %s, Args: %v\n", fn.Name(), fn.Arguments())
	// Output: Function: Integer, Args: [18 120]
}
