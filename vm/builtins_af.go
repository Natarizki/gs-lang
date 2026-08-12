package vm

import (
     "strings"
     "sort"
)

// BuiltinsAF adalah daftar fungsi tambahan yang tersedia lewat import "GS/af"
var BuiltinsAF = []struct {
	Name    string
	Builtin *BuiltinFunction
}{
	{
		"upper",
		&BuiltinFunction{Name: "upper", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "upper() expects 1 argument"}
			}
			s, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "upper() expects a string"}
			}
			return &String{Value: strings.ToUpper(s.Value)}
		}},
	},
	{
		"lower",
		&BuiltinFunction{Name: "lower", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "lower() expects 1 argument"}
			}
			s, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "lower() expects a string"}
			}
			return &String{Value: strings.ToLower(s.Value)}
		}},
	},
	{
		"split",
		&BuiltinFunction{Name: "split", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "split() expects 2 arguments"}
			}
			s, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "split() first argument must be a string"}
			}
			sep, ok := args[1].(*String)
			if !ok {
				return &ErrorValue{Message: "split() second argument must be a string"}
			}
			parts := strings.Split(s.Value, sep.Value)
			elements := make([]Value, len(parts))
			for i, p := range parts {
				elements[i] = &String{Value: p}
			}
			return &List{Elements: elements}
		}},
	},
	{
		"join",
		&BuiltinFunction{Name: "join", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "join() expects 2 arguments"}
			}
			list, ok := args[0].(*List)
			if !ok {
				return &ErrorValue{Message: "join() first argument must be a list"}
			}
			sep, ok := args[1].(*String)
			if !ok {
				return &ErrorValue{Message: "join() second argument must be a string"}
			}
			parts := make([]string, len(list.Elements))
			for i, el := range list.Elements {
				parts[i] = el.Inspect()
			}
			return &String{Value: strings.Join(parts, sep.Value)}
		}},
	},
	{
		"trim",
		&BuiltinFunction{Name: "trim", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "trim() expects 1 argument"}
			}
			s, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "trim() expects a string"}
			}
			return &String{Value: strings.TrimSpace(s.Value)}
		}},
	},
	{
		"replace",
		&BuiltinFunction{Name: "replace", Fn: func(args ...Value) Value {
			if len(args) != 3 {
				return &ErrorValue{Message: "replace() expects 3 arguments"}
			}
			s, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "replace() first argument must be a string"}
			}
			old, ok := args[1].(*String)
			if !ok {
				return &ErrorValue{Message: "replace() second argument must be a string"}
			}
			new, ok := args[2].(*String)
			if !ok {
				return &ErrorValue{Message: "replace() third argument must be a string"}
			}
			return &String{Value: strings.ReplaceAll(s.Value, old.Value, new.Value)}
		}},
	},
	{
		"contains",
		&BuiltinFunction{Name: "contains", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "contains() expects 2 arguments"}
			}
			s, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "contains() first argument must be a string"}
			}
			sub, ok := args[1].(*String)
			if !ok {
				return &ErrorValue{Message: "contains() second argument must be a string"}
			}
			return NativeBoolToBooleanValue(strings.Contains(s.Value, sub.Value))
		}},
	},
	{
		"sort",
		&BuiltinFunction{Name: "sort", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "sort() expects 1 argument"}
			}
			list, ok := args[0].(*List)
			if !ok {
				return &ErrorValue{Message: "sort() expects a list"}
			}
			sorted := make([]Value, len(list.Elements))
			copy(sorted, list.Elements)
			sort.Slice(sorted, func(i, j int) bool {
				return compareValues(sorted[i], sorted[j]) < 0
			})
			return &List{Elements: sorted}
		}},
	},
	{
		"reverse",
		&BuiltinFunction{Name: "reverse", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "reverse() expects 1 argument"}
			}
			list, ok := args[0].(*List)
			if !ok {
				return &ErrorValue{Message: "reverse() expects a list"}
			}
			reversed := make([]Value, len(list.Elements))
			for i, el := range list.Elements {
				reversed[len(list.Elements)-1-i] = el
			}
			return &List{Elements: reversed}
		}},
	},
	{
		"remove",
		&BuiltinFunction{Name: "remove", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "remove() expects 2 arguments"}
			}
			list, ok := args[0].(*List)
			if !ok {
				return &ErrorValue{Message: "remove() first argument must be a list"}
			}
			idx, ok := args[1].(*Int)
			if !ok {
				return &ErrorValue{Message: "remove() second argument must be an integer"}
			}
			i := int(idx.Value)
			if i < 0 || i >= len(list.Elements) {
				return &ErrorValue{Message: "remove() index out of range"}
			}
			newElements := make([]Value, 0, len(list.Elements)-1)
			newElements = append(newElements, list.Elements[:i]...)
			newElements = append(newElements, list.Elements[i+1:]...)
			return &List{Elements: newElements}
		}},
	},
	{
		"index",
		&BuiltinFunction{Name: "index", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "index() expects 2 arguments"}
			}
			list, ok := args[0].(*List)
			if !ok {
				return &ErrorValue{Message: "index() first argument must be a list"}
			}
			for i, el := range list.Elements {
				if compareValues(el, args[1]) == 0 {
					return &Int{Value: int64(i)}
				}
			}
			return &Int{Value: -1}
		}},
	},
	{
		"keys",
		&BuiltinFunction{Name: "keys", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "keys() expects 1 argument"}
			}
			m, ok := args[0].(*Map)
			if !ok {
				return &ErrorValue{Message: "keys() expects a map"}
			}
			elements := make([]Value, 0, len(m.Pairs))
			for _, pair := range m.Pairs {
				elements = append(elements, pair.Key)
			}
			return &List{Elements: elements}
		}},
	},
	{
		"values",
		&BuiltinFunction{Name: "values", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "values() expects 1 argument"}
			}
			m, ok := args[0].(*Map)
			if !ok {
				return &ErrorValue{Message: "values() expects a map"}
			}
			elements := make([]Value, 0, len(m.Pairs))
			for _, pair := range m.Pairs {
				elements = append(elements, pair.Value)
			}
			return &List{Elements: elements}
		}},
	},
	{
		"has",
		&BuiltinFunction{Name: "has", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "has() expects 2 arguments"}
			}
			m, ok := args[0].(*Map)
			if !ok {
				return &ErrorValue{Message: "has() first argument must be a map"}
			}
			_, exists := m.Get(args[1])
			return NativeBoolToBooleanValue(exists)
		}},
	},
	{
		"readFile",
		&BuiltinFunction{Name: "readFile", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "readFile() expects 1 argument"}
			}
			path, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "readFile() expects a string path"}
			}
			data, err := osReadFile(path.Value)
			if err != nil {
				return &ErrorValue{Message: "readFile() error: " + err.Error()}
			}
			return &String{Value: string(data)}
		}},
	},
	{
		"writeFile",
		&BuiltinFunction{Name: "writeFile", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "writeFile() expects 2 arguments"}
			}
			path, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "writeFile() first argument must be a string path"}
			}
			content, ok := args[1].(*String)
			if !ok {
				return &ErrorValue{Message: "writeFile() second argument must be a string"}
			}
			if err := osWriteFile(path.Value, []byte(content.Value)); err != nil {
				return &ErrorValue{Message: "writeFile() error: " + err.Error()}
			}
			return NULL
		}},
	},
	{
		"appendFile",
		&BuiltinFunction{Name: "appendFile", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "appendFile() expects 2 arguments"}
			}
			path, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "appendFile() first argument must be a string path"}
			}
			content, ok := args[1].(*String)
			if !ok {
				return &ErrorValue{Message: "appendFile() second argument must be a string"}
			}
			if err := osAppendFile(path.Value, content.Value); err != nil {
				return &ErrorValue{Message: "appendFile() error: " + err.Error()}
			}
			return NULL
		}},
	},
	{
		"deleteFile",
		&BuiltinFunction{Name: "deleteFile", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "deleteFile() expects 1 argument"}
			}
			path, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "deleteFile() expects a string path"}
			}
			if err := osRemoveFile(path.Value); err != nil {
				return &ErrorValue{Message: "deleteFile() error: " + err.Error()}
			}
			return NULL
		}},
	},
	{
		"now",
		&BuiltinFunction{Name: "now", Fn: func(args ...Value) Value {
			return &String{Value: nowString()}
		}},
	},
	{
		"sleep",
		&BuiltinFunction{Name: "sleep", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "sleep() expects 1 argument"}
			}
			ms, ok := args[0].(*Int)
			if !ok {
				return &ErrorValue{Message: "sleep() expects an integer (milliseconds)"}
			}
			sleepMillis(ms.Value)
			return NULL
		}},
	},
	{
		"format_time",
		&BuiltinFunction{Name: "format_time", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "format_time() expects 1 argument"}
			}
			layout, ok := args[0].(*String)
			if !ok {
				return &ErrorValue{Message: "format_time() expects a string layout"}
			}
			return &String{Value: formatTimeNow(layout.Value)}
		}},
	},
	{
		"random",
		&BuiltinFunction{Name: "random", Fn: func(args ...Value) Value {
			if len(args) != 2 {
				return &ErrorValue{Message: "random() expects 2 arguments (min, max)"}
			}
			minV, ok := args[0].(*Int)
			if !ok {
				return &ErrorValue{Message: "random() first argument must be an integer"}
			}
			maxV, ok := args[1].(*Int)
			if !ok {
				return &ErrorValue{Message: "random() second argument must be an integer"}
			}
			return &Int{Value: randomInRange(minV.Value, maxV.Value)}
		}},
	},
	{
		"assert",
		&BuiltinFunction{Name: "assert", Fn: func(args ...Value) Value {
			if len(args) < 1 {
				return &ErrorValue{Message: "assert() expects at least 1 argument"}
			}
			if !IsTruthy(args[0]) {
				msg := "assertion failed"
				if len(args) > 1 {
					msg = args[1].Inspect()
				}
				return &ErrorValue{Message: msg}
			}
			return NULL
		}},
	},
	{
		"copy",
		&BuiltinFunction{Name: "copy", Fn: func(args ...Value) Value {
			if len(args) != 1 {
				return &ErrorValue{Message: "copy() expects 1 argument"}
			}
			return deepCopyValue(args[0])
		}},
	},
	{
		"exit",
		&BuiltinFunction{Name: "exit", Fn: func(args ...Value) Value {
			code := 0
			if len(args) == 1 {
				if i, ok := args[0].(*Int); ok {
					code = int(i.Value)
				}
			}
			exitProgram(code)
			return NULL
		}},
	},
}

// GetBuiltinAFByName mencari fungsi built-in GS/af berdasarkan nama
func GetBuiltinAFByName(name string) *BuiltinFunction {
	for _, b := range BuiltinsAF {
		if b.Name == name {
			return b.Builtin
		}
	}
	return nil
}

// compareValues membandingkan dua Value secara generik: number dibanding
// secara numerik, string secara leksikal, tipe lain dibanding lewat Inspect()
func compareValues(a, b Value) int {
	af, aIsNum := toFloatValue(a)
	bf, bIsNum := toFloatValue(b)
	if aIsNum && bIsNum {
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}

	as, bs := a.Inspect(), b.Inspect()
	switch {
	case as < bs:
		return -1
	case as > bs:
		return 1
	default:
		return 0
	}
}

func toFloatValue(v Value) (float64, bool) {
	switch v := v.(type) {
	case *Int:
		return float64(v.Value), true
	case *Float:
		return v.Value, true
	default:
		return 0, false
	}
}
