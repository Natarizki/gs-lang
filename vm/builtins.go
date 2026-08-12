package vm

import (
	"bufio"
	"fmt"
	"os"
)

var stdinReader = bufio.NewReader(os.Stdin)

var Builtins = []struct {
	Name    string
	Builtin *BuiltinFunction
}{
	{
		"print",
		&BuiltinFunction{Name: "print", Fn: func(args ...Value) Value {
			parts := make([]interface{}, len(args))
			for i, a := range args {
				parts[i] = a.Inspect()
			}
			fmt.Println(parts...)
			return NULL
		}},
	},
	{
		"len",
		&BuiltinFunction{Name: "len", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "len() expects 1 argument"}
			}
			switch v := args[0].(type) {
			case *String:
				return &Int{Value: int64(len(v.Value))}
			case *List:
				return &Int{Value: int64(len(v.Elements))}
			case *Map:
				return &Int{Value: int64(len(v.Pairs))}
			default:
				return &ErrorValue{Message: "len() not supported for this type"}
			}
		}},
	},
	{
		"range",
		&BuiltinFunction{Name: "range", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "range() expects 1 argument"}
			}
			n, ok := args[0].(*Int)
			if !ok {
				return &ErrorValue{Message: "range() expects an integer"}
			}
			elements := make([]Value, n.Value)
			for i := int64(0); i < n.Value; i++ {
				elements[i] = &Int{Value: i}
			}
			return &List{Elements: elements}
		}},
	},
	{
		"type",
		&BuiltinFunction{Name: "type", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "type() expects 1 argument"}
			}
			return &String{Value: string(args[0].Type())}
		}},
	},
	{
		"input",
		&BuiltinFunction{Name: "input", Fn: func(args ...Value) Value {
			if len(args) == 1 {
				fmt.Print(args[0].Inspect())
			}
			line, _ := stdinReader.ReadString('\n')
			line = trimNewline(line)
			return &String{Value: line}
		}},
	},
	{
		"append",
		&BuiltinFunction{Name: "append", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "append() expects 2 arguments"}
			}
			list, ok := args[0].(*List)
			if !ok {
				return &ErrorValue{Message: "append() first argument must be a list"}
			}
			newElements := make([]Value, len(list.Elements)+1)
			copy(newElements, list.Elements)
			newElements[len(list.Elements)] = args[1]
			return &List{Elements: newElements}
		}},
	},
	{
		"str",
		&BuiltinFunction{Name: "str", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "str() expects 1 argument"}
			}
			return &String{Value: args[0].Inspect()}
		}},
	},
	{
		"int",
		&BuiltinFunction{Name: "int", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "int() expects 1 argument"}
			}
			switch v := args[0].(type) {
			case *Int:
				return v
			case *Float:
				return &Int{Value: int64(v.Value)}
			case *String:
				var i int64
				_, err := fmt.Sscanf(v.Value, "%d", &i)
				if err != nil {
					return &ErrorValue{Message: "could not convert to int"}
				}
				return &Int{Value: i}
			default:
				return &ErrorValue{Message: "int() not supported for this type"}
			}
		}},
	},
	{
		"float",
		&BuiltinFunction{Name: "float", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "float() expects 1 argument"}
			}
			switch v := args[0].(type) {
			case *Float:
				return v
			case *Int:
				return &Float{Value: float64(v.Value)}
			case *String:
				var f float64
				_, err := fmt.Sscanf(v.Value, "%g", &f)
				if err != nil {
					return &ErrorValue{Message: "could not convert to float"}
				}
				return &Float{Value: f}
			default:
				return &ErrorValue{Message: "float() not supported for this type"}
			}
		}},
	},
	{
		"bool",
		&BuiltinFunction{Name: "bool", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "bool() expects 1 argument"}
			}
			return NativeBoolToBooleanValue(IsTruthy(args[0]))
		}},
	},
}

func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func GetBuiltinByName(name string) *BuiltinFunction {
	for _, b := range Builtins {
		if b.Name == name {
			return b.Builtin
		}
	}
	return nil
}
