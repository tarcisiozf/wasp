package instructions_test

import (
	"github.com/tarcisiozf/wasp/internal/binary"
	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/memory"
	"github.com/tarcisiozf/wasp/internal/memory/stack"
)

func createTestContext(body ...[]byte) *execution.Context {
	var bytes []byte
	for _, item := range body {
		bytes = append(bytes, item...)
	}

	st := stack.NewWithCapacity(16)
	locals := stack.NewWithCapacity(16)
	globals := &memory.Global{}

	iter := binary.NewIterator(bytes)

	mem := &execution.Memory{
		Globals: globals,
		Stack:   st,
	}

	return &execution.Context{
		Memory: mem,
		Locals: locals,

		Body: iter,
	}
}
