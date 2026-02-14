package instructions

import (
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/opcodes"
)

var (
	LocalGet = addInstruction(opcodes.LocalGet, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		value := ctx.Locals.At(localIndex)
		ctx.Stack.Push(value)
		return nil
	})
	LocalSet = addInstruction(opcodes.LocalSet, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		ctx.Locals.Set(localIndex, ctx.Stack.Pop())
		return nil
	})
	LocalTee = addInstruction(opcodes.LocalTee, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		value := ctx.Stack.Peek()
		ctx.Locals.Set(localIndex, value)
		return nil
	})

	GlobalGet = addInstruction(opcodes.GlobalGet, func(ctx *execution.Context) error {
		globalIndex := ctx.Body.Varint()
		value := ctx.Globals.Get(globalIndex)
		ctx.Stack.Push(value)
		return nil
	})
	GlobalSet = addInstruction(opcodes.GlobalSet, func(ctx *execution.Context) error {
		globalIndex := ctx.Body.Varint()
		ctx.Globals.Set(globalIndex, ctx.Stack.Pop())
		return nil
	})
)
