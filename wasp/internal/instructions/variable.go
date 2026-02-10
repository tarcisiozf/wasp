package instructions

import (
	"wasp/wasp/internal/execution"
)

var (
	LocalGet = addInstruction(0x20, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		value := ctx.Local[localIndex]
		ctx.Stack.Push(value)
		return nil
	})
	LocalSet = addInstruction(0x21, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		ctx.Local[localIndex] = ctx.Stack.Pop()
		return nil
	})
	LocalTee = addInstruction(0x22, func(ctx *execution.Context) error {
		localIndex := ctx.Body.Varint()
		value := ctx.Stack.Peek()
		ctx.Local[localIndex] = value
		return nil
	})
)
