package instructions_test

import (
	"wasp/wasp/internal/binary"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/memory"
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

	return &execution.Context{
		Stack:   stack,
		Locals:  locals,
		Globals: globals,

		Body: iter,
	}
}
