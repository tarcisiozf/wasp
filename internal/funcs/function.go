package funcs

import (
	"fmt"

	"github.com/tarcisiozf/wasp/internal/execution"
	"github.com/tarcisiozf/wasp/internal/funcs/fnblock"
	"github.com/tarcisiozf/wasp/internal/funcs/fnsig"
	"github.com/tarcisiozf/wasp/internal/instructions"
	opcodes "github.com/tarcisiozf/wasp/internal/opcodes"
)

type Function struct {
	Name      string
	Signature fnsig.Signature
	Locals    []any
	Body      []byte

	// Blocks maps block start positions to their precomputed targets
	Blocks map[int]fnblock.Target

	Index  int
	Offset int
}

func (f *Function) String() string {
	if f.Name != "" {
		return f.Name
	}
	return fmt.Sprintf("func_%d", f.Index)
}

func (f *Function) Call(ctx *execution.Context) error {
	var opcode opcodes.Opcode

	for ctx.Body.HasNext() && !ctx.Paused {
		ctx.Body.SetCheckpoint()
		opcode = ctx.Body.Opcode()
		ix := instructions.Instruction(opcode)
		if ctx.Debug {
			fmt.Printf("\t%08x:\t%04x\t%s\n", f.Offset+ctx.Body.Checkpoint(), opcode, opcodes.Name(opcode))
		}
		if ix.Handler == nil { // TODO: remove before flight
			return fmt.Errorf("unimplemented instruction: %s", opcodes.Name(opcode))
		}
		if err := ix.Handler(ctx); err != nil {
			return fmt.Errorf("error executing instruction %s: %w", opcodes.Name(opcode), err)
		}
		if ctx.FunctionCallRequest >= 0 {
			return nil // pause execution to handle function call
		}
		if ctx.Done {
			return nil // function execution complete
		}
	}
	return nil
}
