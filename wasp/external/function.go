package external

import (
	"fmt"
	"reflect"
)

type Function struct {
	ModuleName string
	FieldName  string
	NumInputs  int
	NumOutputs int
	valueOf    reflect.Value
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

func WrapFunc(moduleName, fieldName string, handler any) (*Function, error) {
	valueOf := reflect.ValueOf(handler)
	typeOf := valueOf.Type()

	if typeOf.Kind() != reflect.Func {
		return nil, fmt.Errorf("handler must be a function")
	}

	return &Function{
		ModuleName: moduleName,
		FieldName:  fieldName,

		NumInputs:  typeOf.NumIn(),
		NumOutputs: typeOf.NumOut(),

		valueOf: valueOf,
	}, nil
}
