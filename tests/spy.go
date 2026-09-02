package tests

import (
	"reflect"
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
		if !valuesEqual(call.args[i], arg) {
			t.Errorf("expected argument %d to be %v (%T), got %v (%T)", i, arg, arg, call.args[i], call.args[i])
		}
	}
}

// valuesEqual compares two values, handling numeric type conversions
func valuesEqual(a, b any) bool {
	// Try direct comparison first
	if a == b {
		return true
	}

	// Handle numeric type conversions
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Check if both are integers
	if aVal.CanInt() && bVal.CanInt() {
		return aVal.Int() == bVal.Int()
	}
	if aVal.CanUint() && bVal.CanUint() {
		return aVal.Uint() == bVal.Uint()
	}
	if aVal.CanFloat() && bVal.CanFloat() {
		return aVal.Float() == bVal.Float()
	}

	return false
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
