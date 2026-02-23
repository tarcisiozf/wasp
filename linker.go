package wasp

import (
	"fmt"
	"reflect"

	"github.com/tarcisiozf/wasp/internal/external"
)

type Linker struct {
	modules map[string]map[string]*external.Function
}

func NewLinker() *Linker {
	return &Linker{
		modules: make(map[string]map[string]*external.Function),
	}
}

func (linker *Linker) Define(moduleName, fieldName string, asExtern any) error {
	if asExtern == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	valueOf := reflect.ValueOf(asExtern)
	typeOf := valueOf.Type()

	if typeOf.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	extFunc, err := external.WrapFunc(moduleName, fieldName, asExtern)
	if err != nil {
		return fmt.Errorf("failed to wrap function: %w", err)
	}

	if _, exists := linker.modules[moduleName]; !exists {
		linker.modules[moduleName] = make(map[string]*external.Function)
	}
	if _, exists := linker.modules[moduleName][fieldName]; exists {
		return fmt.Errorf("function %s.%s is already defined", moduleName, fieldName)
	}
	linker.modules[moduleName][fieldName] = extFunc

	return nil
}

func (linker *Linker) Get(moduleName, fieldName string) (*external.Function, error) {
	module, exists := linker.modules[moduleName]
	if !exists {
		return nil, fmt.Errorf("module %s not found", moduleName)
	}
	fn, exists := module[fieldName]
	if !exists {
		return nil, fmt.Errorf("function %s.%s not found", moduleName, fieldName)
	}
	return fn, nil
}
