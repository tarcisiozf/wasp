package instructions

import (
	"fmt"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

type handler func(ctx *execution.Context) error

type instruction struct {
	Opcode  byte
	Handler handler
}

func (i instruction) String() any {
	return fmt.Sprintf("%s(0x%x)", opcodes.Name(i.Opcode), i.Opcode)
}

var instructions = make([]instruction, 256)

func addInstruction(opcode byte, handler handler) instruction {
	if instructions[opcode].Opcode != 0 {
		panic(fmt.Sprintf("instruction already defined: %v", opcode))
	}
	instructions[opcode].Opcode = opcode
	instructions[opcode].Handler = handler
	return instructions[opcode]
}

func Instruction(opcode byte) instruction {
	return instructions[opcode]
}
