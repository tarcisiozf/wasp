package instructions

import (
	"fmt"
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
		value, _, err := ctx.Globals.Get(globalIndex)
		if err != nil {
			return fmt.Errorf("global get ix: %w", err)
		}
		ctx.Stack.Push(value)
		return nil
	})
	GlobalSet = addInstruction(opcodes.GlobalSet, func(ctx *execution.Context) error {
		globalIndex := ctx.Body.Varint()
		err := ctx.Globals.Set(globalIndex, ctx.Stack.Pop())
		if err != nil {
			return fmt.Errorf("global set ix: %w", err)
		}
		return nil
	})
)
