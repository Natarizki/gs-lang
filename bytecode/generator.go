package bytecode

import (
	"fmt"

	"gs-lang/parser"
)

type Generator struct {
	instructions Instructions
	constants    []interface{}

	symbolTable *SymbolTable

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction

	breakPositions    [][]int
	continuePositions [][]int

	taskFunctionNames []string
	activeTaskPrefix  string

	afImported bool
	foImported bool
}

type EmittedInstruction struct {
	Opcode   Opcode
	Position int
}

func New() *Generator {
	return &Generator{
		instructions: Instructions{},
		constants:    []interface{}{},
		symbolTable:  NewSymbolTable(),
	}
}

type Bytecode struct {
	Instructions Instructions
	Constants    []interface{}
}

func (g *Generator) Bytecode() *Bytecode {
	if !g.lastInstructionIs(OpHalt) {
		g.emit(OpHalt)
	}
	return &Bytecode{Instructions: g.instructions, Constants: g.constants}
}

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol)}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	sym, ok := s.store[name]
	if !ok && s.Outer != nil {
		return s.Outer.Resolve(name)
	}
	return sym, ok
}

func (g *Generator) emit(op Opcode, operands ...int) int {
	ins := Make(op, operands...)
	pos := g.addInstruction(ins)
	g.setLastInstruction(op, pos)
	return pos
}

func (g *Generator) addInstruction(ins []byte) int {
	posNewInstruction := len(g.instructions)
	g.instructions = append(g.instructions, ins...)
	return posNewInstruction
}

func (g *Generator) setLastInstruction(op Opcode, pos int) {
	g.previousInstruction = g.lastInstruction
	g.lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

func (g *Generator) lastInstructionIs(op Opcode) bool {
	if len(g.instructions) == 0 {
		return false
	}
	return g.lastInstruction.Opcode == op
}

func (g *Generator) removeLastPop() {
	g.instructions = g.instructions[:g.lastInstruction.Position]
	g.lastInstruction = g.previousInstruction
}

func (g *Generator) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		g.instructions[pos+i] = newInstruction[i]
	}
}

func (g *Generator) changeOperand(opPos int, operand int) {
	op := Opcode(g.instructions[opPos])
	newInstruction := Make(op, operand)
	g.replaceInstruction(opPos, newInstruction)
}

func (g *Generator) addConstant(obj interface{}) int {
	g.constants = append(g.constants, obj)
	return len(g.constants) - 1
}

func (g *Generator) Compile(node parser.Node) error {
	switch node := node.(type) {

	case *parser.Program:
		for _, stmt := range node.Statements {
			if imp, ok := stmt.(*parser.ImportStatement); ok {
				if imp.Path == "GS/af" {
					g.afImported = true
				}
				if imp.Path == "GS/fo" {
					g.foImported = true
				}
			}
		}

		taskBlocks := []*parser.TaskBlock{}
		for _, stmt := range node.Statements {
			if tb, ok := stmt.(*parser.TaskBlock); ok {
				g.symbolTable.Define("__task_" + tb.ID)
				taskBlocks = append(taskBlocks, tb)
			}
		}

		// pra-daftarkan semua fnc top-level (di luar task) sebagai global
		for _, stmt := range node.Statements {
			if fs, ok := stmt.(*parser.FuncStatement); ok {
				if _, exists := g.symbolTable.Resolve(fs.Name.Value); !exists {
					g.symbolTable.Define(fs.Name.Value)
				}
			}
		}
		for _, stmt := range node.Statements {
			if fs, ok := stmt.(*parser.FuncStatement); ok {
				if err := g.compileFunctionLiteral(fs.Parameters, fs.Body); err != nil {
					return err
				}
				symbol, _ := g.symbolTable.Resolve(fs.Name.Value)
				g.emit(OpSetGlobal, symbol.Index)
			}
		}

		// pra-daftarkan dan compile struct top-level (di luar task) sebagai global
		for _, stmt := range node.Statements {
			if ss, ok := stmt.(*parser.StructStatement); ok {
				if _, exists := g.symbolTable.Resolve(ss.Name.Value); !exists {
					g.symbolTable.Define(ss.Name.Value)
				}
			}
		}
		for _, stmt := range node.Statements {
			if ss, ok := stmt.(*parser.StructStatement); ok {
				fieldNames := make([]string, len(ss.Fields))
				for i, f := range ss.Fields {
					fieldNames[i] = f.Value
				}
				def := &StructDefConstant{Name: ss.Name.Value, Fields: fieldNames}
				idx := g.addConstant(def)
				symbol, _ := g.symbolTable.Resolve(ss.Name.Value)
				g.emit(OpConstant, idx)
				g.emit(OpSetGlobal, symbol.Index)
			}
		}
		// compile method top-level (butuh struct sudah global)
		for _, stmt := range node.Statements {
			if ms, ok := stmt.(*parser.MethodStatement); ok {
				allParams := append([]*parser.Identifier{ms.ReceiverName}, ms.Parameters...)
				if err := g.compileFunctionLiteral(allParams, ms.Body); err != nil {
					return err
				}
				methodIdx := g.addConstant(ms.Name.Value)
				typeSymbol, ok := g.symbolTable.Resolve(ms.ReceiverType.Value)
				if !ok {
					return fmt.Errorf("unknown struct type: %s", ms.ReceiverType.Value)
				}
				g.emit(OpGetGlobal, typeSymbol.Index)
				g.emit(OpAttachMethod, methodIdx)
			}
		}

		for _, tb := range taskBlocks {
			if err := g.compileTaskBlockBody(tb); err != nil {
				return err
			}
		}

		hasStartOrThen := false
		for _, stmt := range node.Statements {
			if _, ok := stmt.(*parser.StartThenStatement); ok {
				hasStartOrThen = true
			}
		}

		if len(taskBlocks) > 0 && !hasStartOrThen {
			// fallback: cuma ada task tanpa start/then eksplisit -> jalankan task pertama otomatis
			if err := g.emitTaskCall(taskBlocks[0].ID); err != nil {
				return err
			}
		}

		for _, stmt := range node.Statements {
			if _, ok := stmt.(*parser.TaskBlock); ok {
				continue
			}
			if _, ok := stmt.(*parser.FuncStatement); ok {
				continue
			}
			if _, ok := stmt.(*parser.StructStatement); ok {
				continue
			}
			if _, ok := stmt.(*parser.MethodStatement); ok {
				continue
			}
			if err := g.Compile(stmt); err != nil {
				return err
			}
		}

	case *parser.ExpressionStatement:
		if err := g.Compile(node.Expression); err != nil {
			return err
		}
		g.emit(OpPop)

	case *parser.LetStatement:
		if err := g.Compile(node.Value); err != nil {
			return err
		}
		symbol := g.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			g.emit(OpSetGlobal, symbol.Index)
		} else {
			g.emit(OpSetLocal, symbol.Index)
		}

	case *parser.BlockStatement:
		for _, stmt := range node.Statements {
			if err := g.Compile(stmt); err != nil {
				return err
			}
		}

	case *parser.IfStatement:
		if err := g.compileIfStatement(node); err != nil {
			return err
		}

	case *parser.ForStatement:
		if err := g.compileForStatement(node); err != nil {
			return err
		}

	case *parser.BreakStatement:
		pos := g.emit(OpJump, 9999)
		if len(g.breakPositions) > 0 {
			last := len(g.breakPositions) - 1
			g.breakPositions[last] = append(g.breakPositions[last], pos)
		}

	case *parser.ContinueStatement:
		pos := g.emit(OpJump, 9999)
		if len(g.continuePositions) > 0 {
			last := len(g.continuePositions) - 1
			g.continuePositions[last] = append(g.continuePositions[last], pos)
		}

        case *parser.ReturnStatement:
		for _, rv := range node.ReturnValues {
			if err := g.Compile(rv); err != nil {
				return err
			}
		}
		g.emit(OpReturnValue, len(node.ReturnValues))

	case *parser.TryStatement:
		if err := g.Compile(node.Call); err != nil {
			return err
		}
		symbol := g.symbolTable.Define(node.Target.Value)
		g.emit(OpTry, 0)
		if symbol.Scope == GlobalScope {
			g.emit(OpSetGlobal, symbol.Index)
		} else {
			g.emit(OpSetLocal, symbol.Index)
		}

        case *parser.SpawnStatement:
		callExpr, ok := node.Call.(*parser.CallExpression)
		if !ok {
			return fmt.Errorf("spawn requires a function call")
		}
		if err := g.Compile(callExpr.Function); err != nil {
			return err
		}
		for _, arg := range callExpr.Arguments {
			if err := g.Compile(arg); err != nil {
				return err
			}
		}
		g.emit(OpSpawn, len(callExpr.Arguments))

	case *parser.WaitStatement:
		g.emit(OpWait)

	case *parser.ErrorStatement:
		if err := g.Compile(node.Message); err != nil {
			return err
		}
		g.emit(OpError)

	case *parser.FuncStatement:
		if err := g.compileFunctionLiteral(node.Parameters, node.Body); err != nil {
			return err
		}
		symbol := g.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			g.emit(OpSetGlobal, symbol.Index)
		} else {
			g.emit(OpSetLocal, symbol.Index)
		}

	case *parser.StartThenStatement:
		for _, id := range node.TaskIDs {
			if err := g.emitTaskCall(id); err != nil {
				return err
			}
		}

	case *parser.StartIfStatement:
		if err := g.Compile(node.Condition); err != nil {
			return err
		}
		jumpNotTruePos := g.emit(OpJumpNotTrue, 9999)

		for _, id := range node.Consequence {
			if err := g.emitTaskCall(id); err != nil {
				return err
			}
		}
		jumpEndPos := g.emit(OpJump, 9999)

		afterConsPos := len(g.instructions)
		g.changeOperand(jumpNotTruePos, afterConsPos)

		for _, id := range node.Alternative {
			if err := g.emitTaskCall(id); err != nil {
				return err
			}
		}

		afterAllPos := len(g.instructions)
		g.changeOperand(jumpEndPos, afterAllPos)

	case *parser.Identifier:
		lookupName := node.Value
		if g.activeTaskPrefix != "" {
			if _, exists := g.symbolTable.Resolve(g.activeTaskPrefix + node.Value); exists {
				lookupName = g.activeTaskPrefix + node.Value
			}
		}
		symbol, ok := g.symbolTable.Resolve(lookupName)
		if !ok {
			if idx, isBuiltin := BuiltinIndex(node.Value); isBuiltin {
				g.emit(OpGetBuiltin, idx)
			} else if idx, isAF := AFBuiltinIndex(node.Value); isAF && g.afImported {
				g.emit(OpGetBuiltinAF, idx)
			} else if idx, isFO := FOBuiltinIndex(node.Value); isFO && g.foImported {
				g.emit(OpGetBuiltinFO, idx)
			} else {
				return fmt.Errorf("undefined variable: %s", node.Value)
			}
		} else if symbol.Scope == GlobalScope {
			g.emit(OpGetGlobal, symbol.Index)
		} else {
			g.emit(OpGetLocal, symbol.Index)
		}

	case *parser.IntegerLiteral:
		g.emit(OpConstant, g.addConstant(node.Value))

	case *parser.FloatLiteral:
		g.emit(OpConstant, g.addConstant(node.Value))

	case *parser.StringLiteral:
		g.emit(OpConstant, g.addConstant(node.Value))

	case *parser.InterpolatedString:
		if len(node.Parts) == 0 {
			g.emit(OpConstant, g.addConstant(""))
			break
		}
		first := true
		for _, part := range node.Parts {
			if part.IsExpr {
				if err := g.Compile(part.Expr); err != nil {
					return err
				}
				g.emit(OpToString)
			} else {
				g.emit(OpConstant, g.addConstant(part.Text))
			}
			if !first {
				g.emit(OpAdd)
			}
			first = false
		}

	case *parser.BooleanLiteral:
		if node.Value {
			g.emit(OpTrue)
		} else {
			g.emit(OpFalse)
		}

	case *parser.NullLiteral:
		g.emit(OpNull)

	case *parser.PrefixExpression:
		if err := g.Compile(node.Right); err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			g.emit(OpMinus)
		case "!":
			g.emit(OpNot)
		default:
			return fmt.Errorf("unknown prefix operator: %s", node.Operator)
		}

	case *parser.InfixExpression:
		if err := g.Compile(node.Left); err != nil {
			return err
		}
		if err := g.Compile(node.Right); err != nil {
			return err
		}
		if err := g.emitInfixOperator(node.Operator); err != nil {
			return err
		}

	case *parser.ListLiteral:
		for _, el := range node.Elements {
			if err := g.Compile(el); err != nil {
				return err
			}
		}
		g.emit(OpList, len(node.Elements))

	case *parser.MapLiteral:
		count := 0
		for k, v := range node.Pairs {
			if err := g.Compile(k); err != nil {
				return err
			}
			if err := g.Compile(v); err != nil {
				return err
			}
			count++
		}
		g.emit(OpMap, count)

	case *parser.IndexExpression:
		if err := g.Compile(node.Left); err != nil {
			return err
		}
		if err := g.Compile(node.Index); err != nil {
			return err
		}
		if node.End != nil {
			if err := g.Compile(node.End); err != nil {
				return err
			}
			g.emit(OpSlice)
		} else {
			g.emit(OpIndex)
		}

	case *parser.CallExpression:
		if dotExpr, ok := node.Function.(*parser.DotExpression); ok {
			// kemungkinan pemanggilan method: instance.method(args)
			if err := g.Compile(dotExpr.Left); err != nil {
				return err
			}
			for _, arg := range node.Arguments {
				if err := g.Compile(arg); err != nil {
					return err
				}
			}
			methodIdx := g.addConstant(dotExpr.Field.Value)
			g.emit(OpCallMethod, methodIdx, len(node.Arguments))
			return nil
		}

		if err := g.Compile(node.Function); err != nil {
			return err
		}
		for _, arg := range node.Arguments {
			if err := g.Compile(arg); err != nil {
				return err
			}
		}
		g.emit(OpCall, len(node.Arguments))

	case *parser.FunctionLiteral:
		if err := g.compileFunctionLiteral(node.Parameters, node.Body); err != nil {
			return err
		}

	case *parser.StructStatement:
		fieldNames := make([]string, len(node.Fields))
		for i, f := range node.Fields {
			fieldNames[i] = f.Value
		}
		def := &StructDefConstant{Name: node.Name.Value, Fields: fieldNames}
		idx := g.addConstant(def)
		symbol := g.symbolTable.Define(node.Name.Value)
		g.emit(OpConstant, idx)
		if symbol.Scope == GlobalScope {
			g.emit(OpSetGlobal, symbol.Index)
		} else {
			g.emit(OpSetLocal, symbol.Index)
		}

        case *parser.MethodStatement:
		// tambahkan receiver sebagai parameter pertama (implisit)
		allParams := append([]*parser.Identifier{node.ReceiverName}, node.Parameters...)
		if err := g.compileFunctionLiteral(allParams, node.Body); err != nil {
			return err
		}

		methodIdx := g.addConstant(node.Name.Value)
		typeSymbol, ok := g.symbolTable.Resolve(node.ReceiverType.Value)
		if !ok {
			return fmt.Errorf("unknown struct type: %s", node.ReceiverType.Value)
		}
		if typeSymbol.Scope == GlobalScope {
			g.emit(OpGetGlobal, typeSymbol.Index)
		} else {
			g.emit(OpGetLocal, typeSymbol.Index)
		}
		// stack sekarang: [compiled_function, struct_def]
		// tukar urutan supaya OpAttachMethod bisa pop dengan urutan yang benar
		g.emit(OpAttachMethod, methodIdx)

	case *parser.StructInstanceLiteral:
		lookupName := node.Name.Value
		if g.activeTaskPrefix != "" {
			if _, exists := g.symbolTable.Resolve(g.activeTaskPrefix + node.Name.Value); exists {
				lookupName = g.activeTaskPrefix + node.Name.Value
			}
		}
		symbol, ok := g.symbolTable.Resolve(lookupName)
		if !ok {
			return fmt.Errorf("unknown struct type: %s", node.Name.Value)
		}
		if symbol.Scope == GlobalScope {
			g.emit(OpGetGlobal, symbol.Index)
		} else {
			g.emit(OpGetLocal, symbol.Index)
		}

		fieldNames := []string{}
		for fname, fexpr := range node.Fields {
			fieldNames = append(fieldNames, fname)
			if err := g.Compile(fexpr); err != nil {
				return err
			}
		}
		namesIdx := g.addConstant(fieldNames)
		g.emit(OpStructInit, namesIdx)

		// simpan instance yang baru dibuat balik ke variabel yang sama
		if symbol.Scope == GlobalScope {
			g.emit(OpSetGlobal, symbol.Index)
		} else {
			g.emit(OpSetLocal, symbol.Index)
		}

	case *parser.UseStatement:
		// no-op untuk sekarang: instance sudah otomatis tersimpan di variabelnya

	case *parser.ImportStatement:
		// no-op: import GS/af dan GS/fo sudah ditangani lewat scan afImported/foImported
		// di awal *parser.Program; import file biasa sudah diresolve sebelum sampai di sini

	case *parser.DotExpression:
		if err := g.Compile(node.Left); err != nil {
			return err
		}
		fieldIdx := g.addConstant(node.Field.Value)
		g.emit(OpGetField, fieldIdx)

	default:
		return fmt.Errorf("compilation not implemented for node type %T", node)
	}

	return nil
}

func (g *Generator) emitInfixOperator(op string) error {
	switch op {
	case "+":
		g.emit(OpAdd)
	case "-":
		g.emit(OpSub)
	case "*":
		g.emit(OpMul)
	case "/":
		g.emit(OpDiv)
	case "%":
		g.emit(OpMod)
	case "**":
		g.emit(OpPow)
	case "==":
		g.emit(OpEqual)
	case "!=":
		g.emit(OpNotEqual)
	case ">":
		g.emit(OpGreaterThan)
	case ">=":
		g.emit(OpGreaterEq)
	case "<":
		g.emit(OpLessThan)
	case "<=":
		g.emit(OpLessEq)
	case "&&":
		g.emit(OpAnd)
	case "||":
		g.emit(OpOr)
	default:
		return fmt.Errorf("unknown infix operator: %s", op)
	}
	return nil
}

func (g *Generator) compileIfStatement(node *parser.IfStatement) error {
	if err := g.Compile(node.Condition); err != nil {
		return err
	}

	jumpNotTruePos := g.emit(OpJumpNotTrue, 9999)

	if err := g.Compile(node.Consequence); err != nil {
		return err
	}
	if g.lastInstructionIs(OpPop) {
		g.removeLastPop()
	}

	jumpEndPositions := []int{}
	jumpEndPositions = append(jumpEndPositions, g.emit(OpJump, 9999))

	afterConsequencePos := len(g.instructions)
	g.changeOperand(jumpNotTruePos, afterConsequencePos)

	for _, clause := range node.ButIfs {
		if err := g.Compile(clause.Condition); err != nil {
			return err
		}
		jnt := g.emit(OpJumpNotTrue, 9999)

		if err := g.Compile(clause.Consequence); err != nil {
			return err
		}
		if g.lastInstructionIs(OpPop) {
			g.removeLastPop()
		}

		jumpEndPositions = append(jumpEndPositions, g.emit(OpJump, 9999))

		afterClausePos := len(g.instructions)
		g.changeOperand(jnt, afterClausePos)
	}

	if node.Alternative != nil {
		if err := g.Compile(node.Alternative); err != nil {
			return err
		}
		if g.lastInstructionIs(OpPop) {
			g.removeLastPop()
		}
	}

	afterAllPos := len(g.instructions)
	for _, pos := range jumpEndPositions {
		g.changeOperand(pos, afterAllPos)
	}

	return nil
}

func (g *Generator) compileForStatement(node *parser.ForStatement) error {
	symbol := g.symbolTable.Define(node.Variable.Value)

	if err := g.Compile(node.Iterable); err != nil {
		return err
	}

	loopStart := len(g.instructions)
	iterJumpPos := g.emit(OpIterNext, 9999)

	if symbol.Scope == GlobalScope {
		g.emit(OpSetGlobal, symbol.Index)
	} else {
		g.emit(OpSetLocal, symbol.Index)
	}

	g.breakPositions = append(g.breakPositions, []int{})
	g.continuePositions = append(g.continuePositions, []int{})

	if err := g.Compile(node.Body); err != nil {
		return err
	}
	if g.lastInstructionIs(OpPop) {
		g.removeLastPop()
	}

	continueTarget := len(g.instructions)
	g.emit(OpJump, loopStart)

	afterLoopPos := len(g.instructions)
	g.changeOperand(iterJumpPos, afterLoopPos)

	lastBreaks := g.breakPositions[len(g.breakPositions)-1]
	for _, pos := range lastBreaks {
		g.changeOperand(pos, afterLoopPos)
	}
	lastContinues := g.continuePositions[len(g.continuePositions)-1]
	for _, pos := range lastContinues {
		g.changeOperand(pos, continueTarget)
	}

	g.breakPositions = g.breakPositions[:len(g.breakPositions)-1]
	g.continuePositions = g.continuePositions[:len(g.continuePositions)-1]

	return nil
}

func (g *Generator) compileFunctionLiteral(params []*parser.Identifier, body *parser.BlockStatement) error {
	enclosedSymbolTable := NewEnclosedSymbolTable(g.symbolTable)

	outerInstructions := g.instructions
	outerLast := g.lastInstruction
	outerPrevious := g.previousInstruction
	outerSymbolTable := g.symbolTable

	g.instructions = Instructions{}
	g.symbolTable = enclosedSymbolTable

	for _, p := range params {
		g.symbolTable.Define(p.Value)
	}

	if err := g.Compile(body); err != nil {
		return err
	}

	if !g.lastInstructionIs(OpReturnValue) && !g.lastInstructionIs(OpReturn) {
		g.emit(OpReturnValue, 0)
	}

	numLocals := g.symbolTable.numDefinitions
	fnInstructions := g.instructions

	g.instructions = outerInstructions
	g.lastInstruction = outerLast
	g.previousInstruction = outerPrevious
	g.symbolTable = outerSymbolTable

	compiledFn := &CompiledFunctionConstant{
		Instructions:  fnInstructions,
		NumLocals:     numLocals,
		NumParameters: len(params),
	}

	g.emit(OpMakeFunction, g.addConstant(compiledFn))
	return nil
}

// compileFunctionLiteralWithNamePrefix mengkompilasi function literal dengan
// prefix aktif, supaya pemanggilan fungsi lain (Identifier) di dalam body-nya
// otomatis di-resolve ke versi yang sudah diberi prefix (kalau ada), sebelum
// jatuh ke pencarian nama biasa
func (g *Generator) compileFunctionLiteralWithNamePrefix(params []*parser.Identifier, body *parser.BlockStatement, prefix string) error {
	previousPrefix := g.activeTaskPrefix
	g.activeTaskPrefix = prefix
	err := g.compileFunctionLiteral(params, body)
	g.activeTaskPrefix = previousPrefix
	return err
}

// compileTaskBlockBody mengkompilasi isi sebuah TaskBlock yang namanya (__task_00X)
// sudah didaftarkan sebelumnya di symbol table (lihat pass 0 di *parser.Program)
func (g *Generator) compileTaskBlockBody(tb *parser.TaskBlock) error {
	prefix := "__task_" + tb.ID + "_"
	wrappedBody := wrapBodyWithMainCallPrefixed(tb.Body, prefix)

	for _, s := range tb.Body.Statements {
		if fs, ok := s.(*parser.FuncStatement); ok {
			uniqueName := prefix + fs.Name.Value
			if _, exists := g.symbolTable.Resolve(uniqueName); !exists {
				g.symbolTable.Define(uniqueName)
			}
		}
		if ss, ok := s.(*parser.StructStatement); ok {
			uniqueName := prefix + ss.Name.Value
			if _, exists := g.symbolTable.Resolve(uniqueName); !exists {
				g.symbolTable.Define(uniqueName)
			}
		}
	}

	if err := g.compileTaskFunctionsAsGlobalsPrefixed(tb.Body, prefix); err != nil {
		return err
	}

	if err := g.compileFunctionLiteralWithNamePrefix([]*parser.Identifier{}, wrappedBody, prefix); err != nil {
		return err
	}
	symbol, ok := g.symbolTable.Resolve("__task_" + tb.ID)
	if !ok {
		return fmt.Errorf("internal error: task __task_%s not pre-registered", tb.ID)
	}
	g.emit(OpSetGlobal, symbol.Index)
	return nil
}

// compileTaskFunctionsAsGlobals mengkompilasi semua fnc level pertama di dalam
// task sebagai fungsi global (namanya sudah didaftarkan sebelumnya), supaya
// fungsi-fungsi dalam task yang sama bisa saling memanggil tanpa closure
func (g *Generator) compileTaskFunctionsAsGlobals(body *parser.BlockStatement) error {
	for _, s := range body.Statements {
		fs, ok := s.(*parser.FuncStatement)
		if !ok {
			continue
		}
		if err := g.compileFunctionLiteral(fs.Parameters, fs.Body); err != nil {
			return err
		}
		symbol, ok := g.symbolTable.Resolve(fs.Name.Value)
		if !ok {
			return fmt.Errorf("internal error: function %s not pre-registered", fs.Name.Value)
		}
		g.emit(OpSetGlobal, symbol.Index)
	}
	return nil
}

// compileTaskFunctionsAsGlobalsPrefixed sama seperti compileTaskFunctionsAsGlobals,
// tapi setiap nama fungsi diberi prefix unik per task supaya "main" di task
// yang berbeda tidak saling menimpa slot global yang sama
func (g *Generator) compileTaskFunctionsAsGlobalsPrefixed(body *parser.BlockStatement, prefix string) error {
	// urutan penting: struct dulu (karena method butuh struct-nya sudah ada)
	for _, s := range body.Statements {
		ss, ok := s.(*parser.StructStatement)
		if !ok {
			continue
		}
		uniqueName := prefix + ss.Name.Value
		fieldNames := make([]string, len(ss.Fields))
		for i, f := range ss.Fields {
			fieldNames[i] = f.Value
		}
		def := &StructDefConstant{Name: ss.Name.Value, Fields: fieldNames}
		idx := g.addConstant(def)
		g.emit(OpConstant, idx)
		symbol, ok := g.symbolTable.Resolve(uniqueName)
		if !ok {
			return fmt.Errorf("internal error: struct %s not pre-registered", uniqueName)
		}
		g.emit(OpSetGlobal, symbol.Index)
	}

	// lalu method (butuh struct sudah global)
	for _, s := range body.Statements {
		ms, ok := s.(*parser.MethodStatement)
		if !ok {
			continue
		}
		allParams := append([]*parser.Identifier{ms.ReceiverName}, ms.Parameters...)
		if err := g.compileFunctionLiteralWithNamePrefix(allParams, ms.Body, prefix); err != nil {
			return err
		}
		methodIdx := g.addConstant(ms.Name.Value)
		typeSymbol, ok := g.symbolTable.Resolve(prefix + ms.ReceiverType.Value)
		if !ok {
			return fmt.Errorf("unknown struct type: %s", ms.ReceiverType.Value)
		}
		g.emit(OpGetGlobal, typeSymbol.Index)
		g.emit(OpAttachMethod, methodIdx)
	}

	// terakhir fungsi biasa
	for _, s := range body.Statements {
		fs, ok := s.(*parser.FuncStatement)
		if !ok {
			continue
		}
		uniqueName := prefix + fs.Name.Value
		if err := g.compileFunctionLiteralWithNamePrefix(fs.Parameters, fs.Body, prefix); err != nil {
			return err
		}
		symbol, ok := g.symbolTable.Resolve(uniqueName)
		if !ok {
			return fmt.Errorf("internal error: function %s not pre-registered", uniqueName)
		}
		g.emit(OpSetGlobal, symbol.Index)
	}

	return nil
}

// wrapBodyWithMainCall menambahkan pemanggilan main() di akhir body task,
// jika task tersebut mendefinisikan fungsi bernama "main"
func wrapBodyWithMainCall(body *parser.BlockStatement) *parser.BlockStatement {
	hasMain := false
	newBody := &parser.BlockStatement{Token: body.Token}

	for _, s := range body.Statements {
		if fs, ok := s.(*parser.FuncStatement); ok {
			if fs.Name.Value == "main" {
				hasMain = true
			}
			continue
		}
		if _, ok := s.(*parser.StructStatement); ok {
			continue
		}
		if _, ok := s.(*parser.MethodStatement); ok {
			continue
		}
		newBody.Statements = append(newBody.Statements, s)
	}

	if hasMain {
		callMain := &parser.ExpressionStatement{
			Expression: &parser.CallExpression{
				Function:  &parser.Identifier{Value: "main"},
				Arguments: []parser.Expression{},
			},
		}
		newBody.Statements = append(newBody.Statements, callMain)
	}

	return newBody
}

// wrapBodyWithMainCallPrefixed sama seperti wrapBodyWithMainCall, tapi
// pemanggilan main() memakai identifier khusus yang akan di-resolve lewat
// activeTaskPrefix generator (lihat case *parser.Identifier)
func wrapBodyWithMainCallPrefixed(body *parser.BlockStatement, prefix string) *parser.BlockStatement {
	hasMain := false
	newBody := &parser.BlockStatement{Token: body.Token}

	for _, s := range body.Statements {
		if fs, ok := s.(*parser.FuncStatement); ok {
			if fs.Name.Value == "main" {
				hasMain = true
			}
			continue
		}
		if _, ok := s.(*parser.StructStatement); ok {
			continue
		}
		if _, ok := s.(*parser.MethodStatement); ok {
			continue
		}
		newBody.Statements = append(newBody.Statements, s)
	}

	if hasMain {
		callMain := &parser.ExpressionStatement{
			Expression: &parser.CallExpression{
				Function:  &parser.Identifier{Value: "main"},
				Arguments: []parser.Expression{},
			},
		}
		newBody.Statements = append(newBody.Statements, callMain)
	}

	return newBody
}

type CompiledFunctionConstant struct {
	Instructions  Instructions
	NumLocals     int
	NumParameters int
}

// StructDefConstant menyimpan definisi struct (nama + urutan field) di constant pool
type StructDefConstant struct {
	Name   string
	Fields []string
}

// emitTaskCall memanggil fungsi task tersembunyi __task_00X berdasarkan ID (misal "001")
func (g *Generator) emitTaskCall(taskID string) error {
	name := "__task_" + taskID
	symbol, ok := g.symbolTable.Resolve(name)
	if !ok {
		return fmt.Errorf("unknown task id::%s", taskID)
	}
	if symbol.Scope == GlobalScope {
		g.emit(OpGetGlobal, symbol.Index)
	} else {
		g.emit(OpGetLocal, symbol.Index)
	}
	g.emit(OpCall, 0)
	g.emit(OpPop)
	return nil
}

var builtinNames = []string{
	"print", "len", "range", "type", "input", "append", "str", "int", "float", "bool",
}

func BuiltinIndex(name string) (int, bool) {
	for i, n := range builtinNames {
		if n == name {
			return i, true
		}
	}
	return 0, false
}
var afBuiltinNames = []string{
	"upper", "lower", "split", "join", "trim", "replace", "contains",
	"sort", "reverse", "remove", "index",
	"keys", "values", "has",
	"readFile", "writeFile", "appendFile", "deleteFile",
	"now", "sleep", "format_time",
	"random", "assert", "copy", "exit",
}

func AFBuiltinIndex(name string) (int, bool) {
	for i, n := range afBuiltinNames {
		if n == name {
			return i, true
		}
	}
	return 0, false
}

var foBuiltinNames = []string{
	"sqrt", "pow", "sin", "cos", "tan", "log", "log10",
	"abs", "floor", "ceil", "round", "max", "min",
}

func FOBuiltinIndex(name string) (int, bool) {
	for i, n := range foBuiltinNames {
		if n == name {
			return i, true
		}
	}
	return 0, false
}
