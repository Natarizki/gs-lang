package bytecode

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

func ReadUint16(ins Instructions, offset int) uint16 {
	return binary.BigEndian.Uint16(ins[offset:])
}

func ReadUint8(ins Instructions, offset int) uint8 {
	return uint8(ins[offset])
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins, offset))
		case 1:
			operands[i] = int(ReadUint8(ins, offset))
		}
		offset += width
	}

	return operands, offset
}

func (ins Instructions) String() string {
	var out strings.Builder

	i := 0
	for i < len(ins) {
		def, ok := definitions[Opcode(ins[i])]
		if !ok {
			fmt.Fprintf(&out, "ERROR: opcode %d tidak dikenal\n", ins[i])
			i++
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: jumlah operand salah untuk %s", def.Name)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: format tidak didukung untuk %s", def.Name)
}
