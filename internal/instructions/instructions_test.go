package instructions_test

import (
	"github.com/tarcisiozf/wasp/internal/binary"
	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/memory"
)

func createTestContext(body ...[]byte) *execution.Context {
	var bytes []byte
	for _, item := range body {
		bytes = append(bytes, item...)
	}

	stack := memory.NewStackWithCapacity[any](16)
	locals := memory.NewStackWithCapacity[any](16)
	globals := &memory.Global{}

	iter := binary.NewIterator(bytes)

	mem := &execution.Memory{
		Globals: globals,
		Stack:   stack,
	}

	return &execution.Context{
		Memory: mem,
		Locals: locals,

		Body: iter,
	}
}
