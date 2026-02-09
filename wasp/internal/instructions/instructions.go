package instructions

import (
	"fmt"
	"wasp/wasp/internal/execution"
)

type Handler func(ctx *execution.Context) error

var instructions = make([]Handler, 256)

func addInstruction(opcode byte, handler Handler) Handler {
	if instructions[opcode] != nil {
		panic(fmt.Sprintf("instruction already defined: %v", opcode))
	}
	instructions[opcode] = handler
	return handler
}

func Instruction(opcode byte) Handler {
	return instructions[opcode]
}
