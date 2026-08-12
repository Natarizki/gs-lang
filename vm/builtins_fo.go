package vm

import "math"

// BuiltinsFO adalah daftar fungsi matematika lanjutan yang tersedia lewat import "GS/fo"
var BuiltinsFO = []struct {
	Name    string
	Builtin *BuiltinFunction
}{
	{
		"sqrt",
		&BuiltinFunction{Name: "sqrt", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "sqrt() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "sqrt() expects a number"}
			}
			return &Float{Value: math.Sqrt(f)}
		}},
	},
	{
		"pow",
		&BuiltinFunction{Name: "pow", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "pow() expects 2 arguments"}
			}
			base, ok1 := toFloatValue(args[0])
			exp, ok2 := toFloatValue(args[1])
			if !ok1 || !ok2 {
				return &ErrorValue{Message: "pow() expects numbers"}
			}
			return &Float{Value: math.Pow(base, exp)}
		}},
	},
	{
		"sin",
		&BuiltinFunction{Name: "sin", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "sin() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "sin() expects a number"}
			}
			return &Float{Value: math.Sin(f)}
		}},
	},
	{
		"cos",
		&BuiltinFunction{Name: "cos", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "cos() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "cos() expects a number"}
			}
			return &Float{Value: math.Cos(f)}
		}},
	},
	{
		"tan",
		&BuiltinFunction{Name: "tan", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "tan() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "tan() expects a number"}
			}
			return &Float{Value: math.Tan(f)}
		}},
	},
	{
		"log",
		&BuiltinFunction{Name: "log", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "log() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "log() expects a number"}
			}
			return &Float{Value: math.Log(f)}
		}},
	},
	{
		"log10",
		&BuiltinFunction{Name: "log10", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "log10() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "log10() expects a number"}
			}
			return &Float{Value: math.Log10(f)}
		}},
	},
	{
		"abs",
		&BuiltinFunction{Name: "abs", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "abs() expects 1 argument"}
			}
			switch v := args[0].(type) {
			case *Int:
				if v.Value < 0 {
					return &Int{Value: -v.Value}
				}
				return v
			case *Float:
				return &Float{Value: math.Abs(v.Value)}
			default:
				return &ErrorValue{Message: "abs() expects a number"}
			}
		}},
	},
	{
		"floor",
		&BuiltinFunction{Name: "floor", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "floor() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "floor() expects a number"}
			}
			return &Int{Value: int64(math.Floor(f))}
		}},
	},
	{
		"ceil",
		&BuiltinFunction{Name: "ceil", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "ceil() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "ceil() expects a number"}
			}
			return &Int{Value: int64(math.Ceil(f))}
		}},
	},
	{
		"round",
		&BuiltinFunction{Name: "round", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "round() expects 1 argument"}
			}
			f, ok := toFloatValue(args[0])
			if !ok {
				return &ErrorValue{Message: "round() expects a number"}
			}
			return &Int{Value: int64(math.Round(f))}
		}},
	},
	{
		"max",
		&BuiltinFunction{Name: "max", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "max() expects 2 arguments"}
			}
			a, ok1 := toFloatValue(args[0])
			b, ok2 := toFloatValue(args[1])
			if !ok1 || !ok2 {
				return &ErrorValue{Message: "max() expects numbers"}
			}
			if a > b {
				return args[0]
			}
			return args[1]
		}},
	},
	{
		"min",
		&BuiltinFunction{Name: "min", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "min() expects 2 arguments"}
			}
			a, ok1 := toFloatValue(args[0])
			b, ok2 := toFloatValue(args[1])
			if !ok1 || !ok2 {
				return &ErrorValue{Message: "min() expects numbers"}
			}
			if a < b {
				return args[0]
			}
			return args[1]
		}},
	},
}

// GetBuiltinFOByName mencari fungsi built-in GS/fo berdasarkan nama
func GetBuiltinFOByName(name string) *BuiltinFunction {
	for _, b := range BuiltinsFO {
		if b.Name == name {
			return b.Builtin
		}
	}
	return nil
}
