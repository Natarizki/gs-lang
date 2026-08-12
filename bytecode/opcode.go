package bytecode

type Opcode byte

const (
	OpConstant Opcode = iota
	OpPop
	OpNull
	OpTrue
	OpFalse

	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPow
	OpMinus

	OpEqual
	OpNotEqual
	OpGreaterThan
	OpGreaterEq
	OpLessThan
	OpLessEq
	OpAnd
	OpOr
	OpNot

	OpSetGlobal
	OpGetGlobal
	OpSetLocal
	OpGetLocal

	OpList
	OpMap
	OpIndex
	OpSlice
	OpSetIndex
    
     // ===== Struct =====
	OpStructInit // buat instance struct baru
	OpGetField   // akses field struct (dot expression)
	OpSetField   // set field struct
	OpAttachMethod // operand: index nama method di constant pool -> attach ke struct def
	OpCallMethod   // operand: [index nama method, jumlah argumen] -> panggil method

	OpCall
	OpReturn
	OpReturnValue
	OpMakeFunction

	OpJump
	OpJumpNotTrue

	OpIterNext
	OpBreak
	OpContinue

	OpError
	OpTry

	OpStringFormat
        OpToString

	OpSpawn
	OpWait

	OpGetBuiltin
        OpGetBuiltinAF
        OpGetBuiltinFO

	OpHalt
)

type Instructions []byte

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpPop:      {"OpPop", []int{}},
	OpNull:     {"OpNull", []int{}},
	OpTrue:     {"OpTrue", []int{}},
	OpFalse:    {"OpFalse", []int{}},

	OpAdd:   {"OpAdd", []int{}},
	OpSub:   {"OpSub", []int{}},
	OpMul:   {"OpMul", []int{}},
	OpDiv:   {"OpDiv", []int{}},
	OpMod:   {"OpMod", []int{}},
	OpPow:   {"OpPow", []int{}},
	OpMinus: {"OpMinus", []int{}},

	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},
	OpGreaterEq:   {"OpGreaterEq", []int{}},
	OpLessThan:    {"OpLessThan", []int{}},
	OpLessEq:      {"OpLessEq", []int{}},
	OpAnd:         {"OpAnd", []int{}},
	OpOr:          {"OpOr", []int{}},
	OpNot:         {"OpNot", []int{}},

	OpSetGlobal: {"OpSetGlobal", []int{2}},
	OpGetGlobal: {"OpGetGlobal", []int{2}},
	OpSetLocal:  {"OpSetLocal", []int{1}},
	OpGetLocal:  {"OpGetLocal", []int{1}},

	OpList:     {"OpList", []int{2}},
	OpMap:      {"OpMap", []int{2}},
	OpIndex:    {"OpIndex", []int{}},
	OpSlice:    {"OpSlice", []int{}},
	OpSetIndex: {"OpSetIndex", []int{}},

	OpStructInit: {"OpStructInit", []int{2}},
	OpGetField:   {"OpGetField", []int{2}},
	OpSetField:   {"OpSetField", []int{2}},
	OpAttachMethod: {"OpAttachMethod", []int{2}},
	OpCallMethod:   {"OpCallMethod", []int{2, 1}},

	OpCall:         {"OpCall", []int{1}},
	OpReturn:       {"OpReturn", []int{}},
	OpReturnValue:  {"OpReturnValue", []int{1}},
	OpMakeFunction: {"OpMakeFunction", []int{2}},

	OpJump:        {"OpJump", []int{2}},
	OpJumpNotTrue: {"OpJumpNotTrue", []int{2}},

	OpIterNext: {"OpIterNext", []int{2}},
	OpBreak:    {"OpBreak", []int{}},
	OpContinue: {"OpContinue", []int{}},

	OpError:  {"OpError", []int{}},
	OpTry:    {"OpTry", []int{2}},

	OpStringFormat: {"OpStringFormat", []int{2}},
        OpToString:     {"OpToString", []int{}},

	OpSpawn: {"OpSpawn", []int{1}},
	OpWait:  {"OpWait", []int{}},

	OpGetBuiltin: {"OpGetBuiltin", []int{1}},
        OpGetBuiltinAF: {"OpGetBuiltinAF", []int{1}},
        OpGetBuiltinFO: {"OpGetBuiltinFO", []int{1}},

	OpHalt: {"OpHalt", []int{}},
}

func Lookup(op Opcode) (*Definition, bool) {
	def, ok := definitions[op]
	return def, ok
}
