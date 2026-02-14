package tests

import (
	"testing"
)

type spyCall struct {
	args []any
}

func (call *spyCall) CalledWith(t *testing.T, args ...any) {
	if len(call.args) != len(args) {
		t.Errorf("expected %d arguments, got %d", len(args), len(call.args))
		return
	}
	for i, arg := range args {
		if call.args[i] != arg {
			t.Errorf("expected argument %d to be %v, got %v", i, arg, call.args[i])
		}
	}
}

type Spy struct {
	calls []*spyCall
}

func (spy *Spy) Called(args ...any) {
	spy.calls = append(spy.calls, &spyCall{args})
}

func (spy *Spy) CalledOnce(t *testing.T) {
	spy.CallCount(t, 1)
}

func (spy *Spy) CallCount(t *testing.T, count int) {
	if len(spy.calls) != count {
		t.Errorf("expected %d calls, got %d", count, len(spy.calls))
	}
}

func (spy *Spy) FirstCall() *spyCall {
	if len(spy.calls) == 0 {
		return nil
	}
	return spy.calls[0]
}

func (spy *Spy) OnCall(index int) *spyCall {
	return spy.calls[index]
}

func NewSpy() *Spy {
	return &Spy{}
}
