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

	results, err := fn(int32(5))
	if err != nil {
		panic(err)
	}

	println(results[0].(int32))
}
