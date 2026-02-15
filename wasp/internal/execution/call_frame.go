package execution

import "fmt"

type Function interface {
	Call(context *Context) error
}

type CallFrame struct {
	FunctionIndex int
	Function      Function
	Context       Context
}

func (cf *CallFrame) Done() bool {
	return cf.Context.Done
}

func (cf *CallFrame) Call() error {
	return cf.Function.Call(&cf.Context)
}

func (cf *CallFrame) Results() ([]any, error) {
	if cf.Context.Done {
		return nil, nil
	}
	return nil, fmt.Errorf("function call not completed yet")
}
