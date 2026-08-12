package vm

import (
	"fmt"

	"gs-lang/bytecode"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

type Frame struct {
	fn          *CompiledFunction
	ip          int
	basePointer int
	isMethod    bool
}

func NewFrame(fn *CompiledFunction, basePointer int) *Frame {
	return &Frame{fn: fn, ip: -1, basePointer: basePointer}
}

func NewMethodFrame(fn *CompiledFunction, basePointer int) *Frame {
	return &Frame{fn: fn, ip: -1, basePointer: basePointer, isMethod: true}
}

func (f *Frame) Instructions() bytecode.Instructions {
	return f.fn.Instructions
}

type VM struct {
	constants []Value

	stack []Value
	sp    int

	globals []Value

	frames      []*Frame
	framesIndex int
}

func New(bc *bytecode.Bytecode) *VM {
	mainFn := &CompiledFunction{Instructions: bc.Instructions}
	mainFrame := NewFrame(mainFn, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   convertConstants(bc.Constants),
		stack:       make([]Value, StackSize),
		sp:          0,
		globals:     make([]Value, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

// rawFieldNames membungkus []string supaya bisa disimpan sebagai Value di constant pool
type rawFieldNames struct {
	Names []string
}

func (r *rawFieldNames) Type() ValueType { return "FIELD_NAMES" }
func (r *rawFieldNames) Inspect() string { return "field names" }

func convertConstants(raw []interface{}) []Value {
	result := make([]Value, len(raw))
	for i, c := range raw {
		switch v := c.(type) {
		case int64:
			result[i] = &Int{Value: v}
		case float64:
			result[i] = &Float{Value: v}
		case string:
			result[i] = &String{Value: v}
		case []string:
			result[i] = &rawFieldNames{Names: v}
		case *bytecode.CompiledFunctionConstant:
			result[i] = &CompiledFunction{
				Instructions:  v.Instructions,
				NumLocals:     v.NumLocals,
				NumParameters: v.NumParameters,
			}
		case *bytecode.StructDefConstant:
			result[i] = &StructDef{Name: v.Name, Fields: v.Fields}
		default:
			result[i] = NULL
		}
	}
	return result
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) push(v Value) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = v
	vm.sp++
	return nil
}

func (vm *VM) pop() Value {
	v := vm.stack[vm.sp-1]
	vm.sp--
	return v
}

func (vm *VM) StackTop() Value {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() Value {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	return vm.run(-1)
}

func (vm *VM) run(stopAtDepth int) error {
	var ip int
	var ins bytecode.Instructions
	var op bytecode.Opcode

	for vm.framesIndex != stopAtDepth && vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		if ip >= len(ins) {
			return fmt.Errorf("internal VM error: ip %d out of bounds (len %d) in frame %d", ip, len(ins), vm.framesIndex)
		}
		op = bytecode.Opcode(ins[ip])

		switch op {
		case bytecode.OpConstant:
			constIndex := bytecode.ReadUint16(ins, ip+1)
			vm.currentFrame().ip += 2
			if err := vm.push(vm.constants[constIndex]); err != nil {
				return err
			}

		case bytecode.OpPop:
			vm.pop()

		case bytecode.OpTrue:
			if err := vm.push(TRUE); err != nil {
				return err
			}
		case bytecode.OpFalse:
			if err := vm.push(FALSE); err != nil {
				return err
			}
		case bytecode.OpNull:
			if err := vm.push(NULL); err != nil {
				return err
			}

		case bytecode.OpAdd, bytecode.OpSub, bytecode.OpMul, bytecode.OpDiv, bytecode.OpMod, bytecode.OpPow:
			if err := vm.executeBinaryOperation(op); err != nil {
				return err
			}

		case bytecode.OpEqual, bytecode.OpNotEqual, bytecode.OpGreaterThan, bytecode.OpGreaterEq,
			bytecode.OpLessThan, bytecode.OpLessEq:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case bytecode.OpAnd, bytecode.OpOr:
			right := vm.pop()
			left := vm.pop()
			var result bool
			if op == bytecode.OpAnd {
				result = IsTruthy(left) && IsTruthy(right)
			} else {
				result = IsTruthy(left) || IsTruthy(right)
			}
			if err := vm.push(NativeBoolToBooleanValue(result)); err != nil {
				return err
			}

		case bytecode.OpMinus:
			operand := vm.pop()
			switch o := operand.(type) {
			case *Int:
				if err := vm.push(&Int{Value: -o.Value}); err != nil {
					return err
				}
			case *Float:
				if err := vm.push(&Float{Value: -o.Value}); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported type for negation: %s", operand.Type())
			}

		case bytecode.OpNot:
			operand := vm.pop()
			if err := vm.push(NativeBoolToBooleanValue(!IsTruthy(operand))); err != nil {
				return err
			}

		case bytecode.OpToString:
			v := vm.pop()
			if err := vm.push(&String{Value: v.Inspect()}); err != nil {
				return err
			}

		case bytecode.OpSetGlobal:
			idx := bytecode.ReadUint16(ins, ip+1)
			vm.currentFrame().ip += 2
			vm.globals[idx] = vm.pop()

		case bytecode.OpGetGlobal:
			idx := bytecode.ReadUint16(ins, ip+1)
			vm.currentFrame().ip += 2
			if err := vm.push(vm.globals[idx]); err != nil {
				return err
			}

		case bytecode.OpSetLocal:
			idx := bytecode.ReadUint8(ins, ip+1)
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(idx)] = vm.pop()

		case bytecode.OpGetLocal:
			idx := bytecode.ReadUint8(ins, ip+1)
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			if err := vm.push(vm.stack[frame.basePointer+int(idx)]); err != nil {
				return err
			}

		case bytecode.OpList:
			numElements := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip += 2
			elements := make([]Value, numElements)
			for i := 0; i < numElements; i++ {
				elements[numElements-1-i] = vm.pop()
			}
			if err := vm.push(&List{Elements: elements}); err != nil {
				return err
			}

		case bytecode.OpMap:
			numPairs := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip += 2
			m := NewMap()
			pairs := make([]MapPair, numPairs)
			for i := 0; i < numPairs; i++ {
				value := vm.pop()
				key := vm.pop()
				pairs[numPairs-1-i] = MapPair{Key: key, Value: value}
			}
			for _, p := range pairs {
				m.Set(p.Key, p.Value)
			}
			if err := vm.push(m); err != nil {
				return err
			}

		case bytecode.OpStructInit:
			namesIdx := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip += 2
			fieldNames := vm.constants[namesIdx].(*rawFieldNames).Names

			values := make([]Value, len(fieldNames))
			for i := len(fieldNames) - 1; i >= 0; i-- {
				values[i] = vm.pop()
			}

			defValue := vm.pop()
			def, ok := defValue.(*StructDef)
			if !ok {
				return fmt.Errorf("cannot initialize non-struct type")
			}

			instance := &StructInstance{Def: def, Values: make(map[string]Value)}
			for i, fname := range fieldNames {
				instance.Values[fname] = values[i]
			}
			if err := vm.push(instance); err != nil {
				return err
			}

		case bytecode.OpGetField:
			fieldIdx := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip += 2
			fieldName := vm.constants[fieldIdx].(*String).Value

			left := vm.pop()
			instance, ok := left.(*StructInstance)
			if !ok {
				return fmt.Errorf("dot access not supported for type %s", left.Type())
			}
			val, ok := instance.Values[fieldName]
			if !ok {
				val = NULL
			}
			if err := vm.push(val); err != nil {
				return err
			}

                case bytecode.OpAttachMethod:
			methodIdx := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip += 2
			methodName := vm.constants[methodIdx].(*String).Value

			structDefValue := vm.pop()
			fn := vm.pop()

			def, ok := structDefValue.(*StructDef)
			if !ok {
				return fmt.Errorf("cannot attach method to non-struct type")
			}
			if def.Methods == nil {
				def.Methods = make(map[string]Value)
			}
			def.Methods[methodName] = fn

                case bytecode.OpCallMethod:
			methodIdx := int(bytecode.ReadUint16(ins, ip+1))
			numArgs := int(bytecode.ReadUint8(ins, ip+3))
			vm.currentFrame().ip += 3
			methodName := vm.constants[methodIdx].(*String).Value

			// instance ada di stack posisi: sp - numArgs - 1
			instanceValue := vm.stack[vm.sp-numArgs-1]
			instance, ok := instanceValue.(*StructInstance)
			if !ok {
				return fmt.Errorf("cannot call method %q on non-struct value", methodName)
			}

			methodFn, ok := instance.Def.Methods[methodName]
			if !ok {
				return fmt.Errorf("undefined method %q on struct %s", methodName, instance.Def.Name)
			}

			// instance sudah ada di posisi bawah stack sebagai argumen pertama
			// (receiver), sisa argumen lain sudah di atasnya, jadi total argumen
			// buat callFunction adalah numArgs + 1 (termasuk receiver)
			switch fn := methodFn.(type) {
			case *CompiledFunction:
				if numArgs+1 != fn.NumParameters {
					return fmt.Errorf("wrong number of arguments for method %q: expected %d, got %d",
						methodName, fn.NumParameters-1, numArgs)
				}
				frame := NewMethodFrame(fn, vm.sp-numArgs-1)
				vm.pushFrame(frame)
				vm.sp = frame.basePointer + fn.NumLocals
			default:
				return fmt.Errorf("method %q is not callable", methodName)
			}

		case bytecode.OpIndex:
			index := vm.pop()
			left := vm.pop()
			if err := vm.executeIndex(left, index); err != nil {
				return err
			}

		case bytecode.OpSlice:
			end := vm.pop()
			start := vm.pop()
			left := vm.pop()
			if err := vm.executeSlice(left, start, end); err != nil {
				return err
			}

		case bytecode.OpJump:
			pos := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip = pos - 1

		case bytecode.OpJumpNotTrue:
			pos := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !IsTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case bytecode.OpIterNext:
			jumpPos := int(bytecode.ReadUint16(ins, ip+1))
			vm.currentFrame().ip += 2
			if err := vm.executeIterNext(jumpPos); err != nil {
				return err
			}

		case bytecode.OpCall:
			numArgs := int(bytecode.ReadUint8(ins, ip+1))
			vm.currentFrame().ip += 1
			if err := vm.callFunction(numArgs); err != nil {
				return err
			}

		case bytecode.OpReturnValue:
			numValues := int(bytecode.ReadUint8(ins, ip+1))
			var retVal Value = NULL
			if numValues > 0 {
				retVal = vm.pop()
				for i := 1; i < numValues; i++ {
					vm.pop()
				}
			}
			frame := vm.popFrame()
			if frame.isMethod {
				vm.sp = frame.basePointer
			} else {
				vm.sp = frame.basePointer - 1
			}
			if err := vm.push(retVal); err != nil {
				return err
			}

		case bytecode.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			if err := vm.push(NULL); err != nil {
				return err
			}

		case bytecode.OpError:
                        msg := vm.pop()
                        errVal := &ErrorValue{Message: msg.Inspect()}
                        frame := vm.popFrame()
                        vm.sp = frame.basePointer - 1
                        if err := vm.push(errVal); err != nil {
                                return err
                        }

                case bytecode.OpTry:
                        vm.currentFrame().ip += 2
                        top := vm.pop()
                        if _, isErr := top.(*ErrorValue); isErr {
                                if err := vm.push(NULL); err != nil {
                                        return err
                                }
                        } else {
                                if err := vm.push(top); err != nil {
                                        return err
                                }
                        }

		case bytecode.OpMakeFunction:
			idx := bytecode.ReadUint16(ins, ip+1)
			vm.currentFrame().ip += 2
			fn := vm.constants[idx]
			if err := vm.push(fn); err != nil {
				return err
			}

		case bytecode.OpGetBuiltin:
			idx := bytecode.ReadUint8(ins, ip+1)
			vm.currentFrame().ip += 1
			builtin := Builtins[idx].Builtin
			if err := vm.push(builtin); err != nil {
				return err
			}

		case bytecode.OpGetBuiltinAF:
			idx := bytecode.ReadUint8(ins, ip+1)
			vm.currentFrame().ip += 1
			builtin := BuiltinsAF[idx].Builtin
			if err := vm.push(builtin); err != nil {
				return err
			}

		case bytecode.OpGetBuiltinFO:
			idx := bytecode.ReadUint8(ins, ip+1)
			vm.currentFrame().ip += 1
			builtin := BuiltinsFO[idx].Builtin
			if err := vm.push(builtin); err != nil {
				return err
			}

                case bytecode.OpSpawn:
			numArgs := int(bytecode.ReadUint8(ins, ip+1))
			vm.currentFrame().ip += 1
			depthBeforeCall := vm.framesIndex
			if err := vm.callFunction(numArgs); err != nil {
				return err
			}
			// jalankan sampai frame spawn ini selesai (versi sekuensial sementara)
			if err := vm.run(depthBeforeCall); err != nil {
				return err
			}
			vm.pop() // buang hasil return spawn

		case bytecode.OpWait:
			// no-op untuk versi sekuensial (semua spawn sudah selesai duluan)

		case bytecode.OpHalt:
                        return nil

                default:
                        return fmt.Errorf("unknown opcode: %d", op)
                }
        }

        if vm.sp > 0 {
                if errVal, ok := vm.stack[vm.sp-1].(*ErrorValue); ok {
                        return fmt.Errorf("%s", errVal.Message)
                }
        }

        return nil
}

func (vm *VM) executeBinaryOperation(op bytecode.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftInt, leftIsInt := left.(*Int)
	rightInt, rightIsInt := right.(*Int)

	if leftIsInt && rightIsInt {
		return vm.executeBinaryIntOperation(op, leftInt, rightInt)
	}

	leftFloat, leftIsFloatOK := toFloat(left)
	rightFloat, rightIsFloatOK := toFloat(right)
	if leftIsFloatOK && rightIsFloatOK {
		return vm.executeBinaryFloatOperation(op, leftFloat, rightFloat)
	}

	if op == bytecode.OpAdd {
		leftStr, leftIsStr := left.(*String)
		rightStr, rightIsStr := right.(*String)
		if leftIsStr && rightIsStr {
			return vm.push(&String{Value: leftStr.Value + rightStr.Value})
		}
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", left.Type(), right.Type())
}

func toFloat(v Value) (float64, bool) {
	switch v := v.(type) {
	case *Float:
		return v.Value, true
	case *Int:
		return float64(v.Value), true
	default:
		return 0, false
	}
}

func (vm *VM) executeBinaryIntOperation(op bytecode.Opcode, left, right *Int) error {
	var result int64
	switch op {
	case bytecode.OpAdd:
		result = left.Value + right.Value
	case bytecode.OpSub:
		result = left.Value - right.Value
	case bytecode.OpMul:
		result = left.Value * right.Value
	case bytecode.OpDiv:
		if right.Value == 0 {
			return fmt.Errorf("division by zero")
		}
		result = left.Value / right.Value
	case bytecode.OpMod:
		if right.Value == 0 {
			return fmt.Errorf("division by zero")
		}
		result = left.Value % right.Value
	case bytecode.OpPow:
		result = intPow(left.Value, right.Value)
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
	return vm.push(&Int{Value: result})
}

func intPow(base, exp int64) int64 {
	result := int64(1)
	for i := int64(0); i < exp; i++ {
		result *= base
	}
	return result
}

func (vm *VM) executeBinaryFloatOperation(op bytecode.Opcode, left, right float64) error {
	var result float64
	switch op {
	case bytecode.OpAdd:
		result = left + right
	case bytecode.OpSub:
		result = left - right
	case bytecode.OpMul:
		result = left * right
	case bytecode.OpDiv:
		if right == 0 {
			return fmt.Errorf("division by zero")
		}
		result = left / right
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}
	return vm.push(&Float{Value: result})
}

func (vm *VM) executeComparison(op bytecode.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftF, leftOK := toFloat(left)
	rightF, rightOK := toFloat(right)

	if leftOK && rightOK {
		var result bool
		switch op {
		case bytecode.OpEqual:
			result = leftF == rightF
		case bytecode.OpNotEqual:
			result = leftF != rightF
		case bytecode.OpGreaterThan:
			result = leftF > rightF
		case bytecode.OpGreaterEq:
			result = leftF >= rightF
		case bytecode.OpLessThan:
			result = leftF < rightF
		case bytecode.OpLessEq:
			result = leftF <= rightF
		}
		return vm.push(NativeBoolToBooleanValue(result))
	}

	var result bool
	switch op {
	case bytecode.OpEqual:
		result = left.Inspect() == right.Inspect() && left.Type() == right.Type()
	case bytecode.OpNotEqual:
		result = !(left.Inspect() == right.Inspect() && left.Type() == right.Type())
	default:
		return fmt.Errorf("unsupported comparison for types %s %s", left.Type(), right.Type())
	}
	return vm.push(NativeBoolToBooleanValue(result))
}

func (vm *VM) executeIndex(left, index Value) error {
	switch l := left.(type) {
	case *List:
		idx, ok := index.(*Int)
		if !ok {
			return fmt.Errorf("list index must be an integer")
		}
		if idx.Value < 0 || idx.Value >= int64(len(l.Elements)) {
			return vm.push(NULL)
		}
		return vm.push(l.Elements[idx.Value])
	case *Map:
		val, ok := l.Get(index)
		if !ok {
			return vm.push(NULL)
		}
		return vm.push(val)
	default:
		return fmt.Errorf("index operator not supported for type %s", left.Type())
	}
}

// executeSlice melakukan operasi slicing pada List (misal angka[1:3])
func (vm *VM) executeSlice(left Value, startVal Value, endVal Value) error {
	list, ok := left.(*List)
	if !ok {
		return fmt.Errorf("slice operator only supported for lists, got %s", left.Type())
	}

	startInt, ok := startVal.(*Int)
	if !ok {
		return fmt.Errorf("slice start index must be an integer")
	}
	endInt, ok := endVal.(*Int)
	if !ok {
		return fmt.Errorf("slice end index must be an integer")
	}

	start := int(startInt.Value)
	end := int(endInt.Value)
	length := len(list.Elements)

	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		start = end
	}

	sliced := make([]Value, end-start)
	copy(sliced, list.Elements[start:end])

	return vm.push(&List{Elements: sliced})
}

func (vm *VM) executeIterNext(jumpPos int) error {
	top := vm.stack[vm.sp-1]
	list, ok := top.(*List)
	if !ok {
		return fmt.Errorf("for-loop can only iterate over a list")
	}

	if list.iterIndex >= len(list.Elements) {
		vm.pop() // buang list, iterasi selesai
		vm.currentFrame().ip = jumpPos - 1
		return nil
	}

	elem := list.Elements[list.iterIndex]
	list.iterIndex++
	return vm.push(elem)
}

func (vm *VM) callFunction(numArgs int) error {
	fnValue := vm.stack[vm.sp-1-numArgs]

	switch fn := fnValue.(type) {
	case *CompiledFunction:
		if numArgs != fn.NumParameters {
			return fmt.Errorf("wrong number of arguments: expected %d, got %d", fn.NumParameters, numArgs)
		}
		frame := NewFrame(fn, vm.sp-numArgs)
		vm.pushFrame(frame)
		vm.sp = frame.basePointer + fn.NumLocals
		return nil
	case *BuiltinFunction:
		args := make([]Value, numArgs)
		for i := 0; i < numArgs; i++ {
			args[numArgs-1-i] = vm.pop()
		}
		vm.pop()
		result := fn.Fn(args...)
		if result == nil {
			result = NULL
		}
		return vm.push(result)
	default:
		return fmt.Errorf("calling non-function value: %s", fnValue.Type())
	}
}
