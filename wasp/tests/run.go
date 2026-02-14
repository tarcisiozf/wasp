package tests

import (
	"fmt"
	"testing"
	"wasp/wasp"
	"wasp/wasp/internal/module"
)

type ExecutionResults struct {
	Module   *module.Module
	Instance *wasp.Instance
	Results  []any
}

func (env *Environment) RunWat(t *testing.T, wat string) (Build, ExecutionResults, error) {
	build, err := BuildWat(t, wat)
	if err != nil {
		return Build{}, ExecutionResults{}, fmt.Errorf("failed to build wat: %w", err)
	}
	execution, err := env.RunWasm(build.Wasm)
	if err != nil {
		return Build{}, ExecutionResults{}, fmt.Errorf("failed to run wasm: %w", err)
	}
	return build, execution, nil
}

func (env *Environment) RunWasm(wasm []byte) (ExecutionResults, error) {
	mod, err := wasp.NewModule(wasm)
	if err != nil {
		return ExecutionResults{}, fmt.Errorf("failed to load module: %w", err)
	}

	fn, err := mod.StartFunction()
	if err != nil {
		return ExecutionResults{}, fmt.Errorf("failed to get start function: %w", err)
	}

	instance, err := wasp.NewInstance(mod, env.instanceOptions...)
	if err != nil {
		return ExecutionResults{}, fmt.Errorf("failed to create instance: %w", err)
	}

	results, err := instance.Call(fn)
	if err != nil {
		return ExecutionResults{}, fmt.Errorf("failed to execute start function: %w", err)
	}

	return ExecutionResults{
		Module:   mod,
		Instance: instance,
		Results:  results,
	}, nil
}
