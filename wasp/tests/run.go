package tests

import (
	"fmt"
	"wasp/wasp"
	"wasp/wasp/internal/execution"
	"wasp/wasp/internal/module"
)

type ExecutionResults struct {
	Module   *module.Module
	Instance *wasp.Instance
}

func (r *ExecutionResults) RunExport(name string, args ...any) (*execution.CallFrame, []any, error) {
	fn, err := r.Module.GetExportedFunction(name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get export function: %w", err)
	}
	return r.run(fn, args...)
}

func (r *ExecutionResults) RunStart() (*execution.CallFrame, error) {
	fn, err := r.Module.StartFunction()
	if err != nil {
		return nil, fmt.Errorf("failed to get start function: %w", err)
	}
	ctx, _, err := r.run(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to execute start function: %w", err)
	}
	return ctx, nil
}

func (r *ExecutionResults) run(index int, args ...any) (*execution.CallFrame, []any, error) {
	callFrame, err := r.Instance.Call(index, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute start function: %w", err)
	}
	if err := r.Instance.Tick(); err != nil {
		return nil, nil, fmt.Errorf("failed to tick instance: %w", err)
	}
	results, err := callFrame.Results()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get results: %w", err)
	}
	return callFrame, results, nil
}

func (env *Environment) CreateInstance(wasm []byte) (ExecutionResults, error) {
	mod, err := wasp.NewModule(wasm)
	if err != nil {
		return ExecutionResults{}, fmt.Errorf("failed to load module: %w", err)
	}

	instance, err := wasp.NewInstance(mod, env.instanceOptions...)
	if err != nil {
		return ExecutionResults{}, fmt.Errorf("failed to create instance: %w", err)
	}

	//fn, err := mod.StartFunction()
	//if err != nil {
	//	return ExecutionResults{}, fmt.Errorf("failed to get start function: %w", err)
	//}
	//

	return ExecutionResults{
		Module:   mod,
		Instance: instance,
	}, nil
}
