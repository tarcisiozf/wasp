package main

import (
	"bytes"
	"os"
	"wasp/wasp"
)

func main() {
	wasm, err := os.ReadFile("math.wasm")
	if err != nil {
		panic(err)
	}

	module, err := wasp.NewModule(wasm)
	if err != nil {
		panic(err)
	}

	store := wasp.NewStore(module)

	funcref, err := module.GetExportedFunction("square")
	if err != nil {
		panic(err)
	}

	instance, err := wasp.NewInstance(module, store)
	if err != nil {
		panic(err)
	}

	_, err = instance.Call(funcref, 5)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = wasp.Foo(&buf, store, instance)
	if err != nil {
		panic(err)
	}

	println("Result:", buf.Len())
}
