package instructions

import (
	"fmt"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

type handler func(ctx *execution.Context) error

type instruction struct {
	Opcode  opcodes.Opcode
	Handler handler
}

var instructions = make([]instruction, 256)
var extensions = make([]instruction, 20)

func addInstruction(opcode opcodes.Opcode, handler handler) instruction {
	if opcode > 0xFF {
		return addInstructionToList(extensions, opcode&0xFF, opcode, handler)
	}
	return addInstructionToList(instructions, opcode, opcode, handler)
}

func addInstructionToList(list []instruction, index uint16, opcode opcodes.Opcode, handler handler) instruction {
	if list[index].Opcode != 0 {
		panic(fmt.Sprintf("instruction already defined: %v", opcode))
	}
	list[index].Opcode = opcode
	list[index].Handler = handler
	return list[index]
}

func Instruction(opcode opcodes.Opcode) instruction {
	if opcode > 0xFF {
		return extensions[opcode&0xFF]
	}
	return instructions[opcode]
}
