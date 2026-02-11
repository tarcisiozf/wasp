package external

import (
	"fmt"
	"reflect"
	"wasp/wasp/internal/funcs"
	"wasp/wasp/internal/types"
)

type Function struct {
	moduleName string
	fieldName  string

	numInputs  int
	numOutputs int

	valueOf reflect.Value
	typeOf  reflect.Type
}

func (f *Function) Call(params []any) ([]any, error) {
	paramValues := make([]reflect.Value, len(params))
	for i, param := range params {
		paramValues[i] = reflect.ValueOf(param)
	}
	results := f.valueOf.Call(paramValues)
	anyResults := make([]any, len(results))
	for i, result := range results {
		anyResults[i] = result.Interface()
	}
	return anyResults, nil
}

func (f *Function) CheckSignatureCompatibility(signature funcs.Signature) error {
	numInputs := f.typeOf.NumIn()
	numExpected := len(signature.Params)

	if f.typeOf.IsVariadic() {
		// Variadic: last param is a slice, can accept 0+ args of that type
		if numExpected < numInputs-1 {
			return fmt.Errorf("function %s.%s requires at least %d parameters, but signature has %d", f.moduleName, f.fieldName, numInputs-1, numExpected)
		}
		// Check fixed parameters
		for i := 0; i < numInputs-1; i++ {
			in := f.typeOf.In(i)
			t := types.ForCode(signature.Params[i])
			if !isTypeCompatible(in, t) {
				return fmt.Errorf("parameter %d of function %s.%s has type %s, but signature expects %s", i, f.moduleName, f.fieldName, in, signature.Params[i])
			}
		}
		// Check variadic parameters against the slice element type
		variadicType := f.typeOf.In(numInputs - 1).Elem()
		for i := numInputs - 1; i < numExpected; i++ {
			t := types.ForCode(signature.Params[i])
			if !isTypeCompatible(variadicType, t) {
				return fmt.Errorf("variadic parameter %d of function %s.%s has type %s, but signature expects %s", i, f.moduleName, f.fieldName, variadicType, signature.Params[i])
			}
		}
	} else {
		if numInputs != numExpected {
			return fmt.Errorf("function %s.%s has %d parameters, but signature expects %d", f.moduleName, f.fieldName, numInputs, numExpected)
		}
		for i := 0; i < numInputs; i++ {
			in := f.typeOf.In(i)
			t := types.ForCode(signature.Params[i])
			if !isTypeCompatible(in, t) {
				return fmt.Errorf("parameter %d of function %s.%s has type %s, but signature expects %s", i, f.moduleName, f.fieldName, in, signature.Params[i])
			}
		}
	}

	// Check outputs
	numOutputs := f.typeOf.NumOut()
	if numOutputs != len(signature.Results) {
		return fmt.Errorf("function %s.%s has %d return values, but signature expects %d", f.moduleName, f.fieldName, numOutputs, len(signature.Results))
	}
	for i := 0; i < numOutputs; i++ {
		out := f.typeOf.Out(i)
		t := types.ForCode(signature.Results[i])
		if !isTypeCompatible(out, t) {
			return fmt.Errorf("return value %d of function %s.%s has type %s, but signature expects %s", i, f.moduleName, f.fieldName, out, signature.Results[i])
		}
	}

	return nil
}

func isTypeCompatible(a reflect.Type, b types.Type) bool {
	return a.Kind() == b.Kind() || a.Kind() == reflect.Interface
}

func (f *Function) NumInputs() int {
	return f.numInputs
}

func (f *Function) ModuleName() string {
	return f.moduleName
}

func (f *Function) FieldName() string {
	return f.fieldName
}

func WrapFunc(moduleName, fieldName string, handler any) (*Function, error) {
	valueOf := reflect.ValueOf(handler)
	typeOf := valueOf.Type()

	if typeOf.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler must be a function")
	}

	return &Function{
		moduleName: moduleName,
		fieldName:  fieldName,

		numInputs:  typeOf.NumIn(),
		numOutputs: typeOf.NumOut(),

		valueOf: valueOf,
		typeOf:  typeOf,
	}, nil
}
