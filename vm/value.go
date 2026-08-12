package vm

import (
	"fmt"
	"strings"

	"gs-lang/bytecode"
)

type ValueType string

const (
	INT_VALUE      ValueType = "INT"
	FLOAT_VALUE    ValueType = "FLOAT"
	STRING_VALUE   ValueType = "STRING"
	BOOL_VALUE     ValueType = "BOOL"
	NULL_VALUE     ValueType = "NULL"
	LIST_VALUE     ValueType = "LIST"
	MAP_VALUE      ValueType = "MAP"
	STRUCT_VALUE   ValueType = "STRUCT"
	FUNCTION_VALUE ValueType = "FUNCTION"
	BUILTIN_VALUE  ValueType = "BUILTIN"
	ERROR_VALUE    ValueType = "ERROR"
)

type Value interface {
	Type() ValueType
	Inspect() string
}

type Int struct {
	Value int64
}

func (i *Int) Type() ValueType { return INT_VALUE }
func (i *Int) Inspect() string { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (f *Float) Type() ValueType { return FLOAT_VALUE }
func (f *Float) Inspect() string { return fmt.Sprintf("%g", f.Value) }

type String struct {
	Value string
}

func (s *String) Type() ValueType { return STRING_VALUE }
func (s *String) Inspect() string { return s.Value }

type Bool struct {
	Value bool
}

func (b *Bool) Type() ValueType { return BOOL_VALUE }
func (b *Bool) Inspect() string { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

func (n *Null) Type() ValueType { return NULL_VALUE }
func (n *Null) Inspect() string { return "null" }

type List struct {
	Elements []Value
        iterIndex int
}

func (l *List) Type() ValueType { return LIST_VALUE }
func (l *List) Inspect() string {
	var parts []string
	for _, e := range l.Elements {
		parts = append(parts, e.Inspect())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

type MapPair struct {
	Key   Value
	Value Value
}

type Map struct {
	Pairs map[string]MapPair
}

func NewMap() *Map {
	return &Map{Pairs: make(map[string]MapPair)}
}

func (m *Map) Type() ValueType { return MAP_VALUE }
func (m *Map) Inspect() string {
	var parts []string
	for _, pair := range m.Pairs {
		parts = append(parts, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (m *Map) Set(key Value, value Value) {
	m.Pairs[key.Inspect()] = MapPair{Key: key, Value: value}
}

func (m *Map) Get(key Value) (Value, bool) {
	pair, ok := m.Pairs[key.Inspect()]
	if !ok {
		return nil, false
	}
	return pair.Value, true
}

type StructDef struct {
	Name    string
	Fields  []string
	Methods map[string]Value
}

func (sd *StructDef) Type() ValueType { return STRUCT_VALUE }
func (sd *StructDef) Inspect() string { return "struct " + sd.Name }

type StructInstance struct {
	Def    *StructDef
	Values map[string]Value
}

func (si *StructInstance) Type() ValueType { return STRUCT_VALUE }
func (si *StructInstance) Inspect() string {
	var parts []string
	for _, f := range si.Def.Fields {
		v := si.Values[f]
		if v == nil {
			v = &Null{}
		}
		parts = append(parts, fmt.Sprintf("%s: %s", f, v.Inspect()))
	}
	return si.Def.Name + " {" + strings.Join(parts, ", ") + "}"
}

type CompiledFunction struct {
	Instructions  bytecode.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() ValueType { return FUNCTION_VALUE }
func (cf *CompiledFunction) Inspect() string { return "function" }

type BuiltinFunction struct {
	Name string
	Fn   func(args ...Value) Value
}

func (bf *BuiltinFunction) Type() ValueType { return BUILTIN_VALUE }
func (bf *BuiltinFunction) Inspect() string { return "builtin function: " + bf.Name }

type ErrorValue struct {
	Message string
}

func (e *ErrorValue) Type() ValueType { return ERROR_VALUE }
func (e *ErrorValue) Inspect() string { return "ERROR: " + e.Message }

var (
	TRUE  = &Bool{Value: true}
	FALSE = &Bool{Value: false}
	NULL  = &Null{}
)

func NativeBoolToBooleanValue(b bool) *Bool {
	if b {
		return TRUE
	}
	return FALSE
}

func IsTruthy(v Value) bool {
	switch v := v.(type) {
	case *Bool:
		return v.Value
	case *Null:
		return false
	case *Int:
		return v.Value != 0
	case *Float:
		return v.Value != 0
	case *String:
		return v.Value != ""
	default:
		return true
	}
}
