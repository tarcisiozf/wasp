package main

import (
	"wasp/wasp"
)

func main() {
	module, err := wasp.NewModuleFromFile("math.wasm")
	if err != nil {
		panic(err)
	}

	fn, err := module.GetExportedFunction("square")
	if err != nil {
		panic(err)
	}

	_ = fn
}
