package main

import (
	"fmt"
	"wasp/wasp"
	"wasp/wasp/external"
)

func main() {
	//module, err := wasp.NewModuleFromFile("math.wasm")
	//if err != nil {
	//	panic(err)
	//}
	//
	//fn, err := module.GetExportedFunction("square")
	//if err != nil {
	//	panic(err)
	//}
	//
	//results, err := fn(int32(5))
	//if err != nil {
	//	panic(err)
	//}
	//
	//println(results[0].(int32))

	consoleLog, err := external.WrapFunc("console", "log", func(args ...any) {
		fmt.Println(args...)
	})
	if err != nil {
		panic(err)
	}

	{
		module, err := wasp.NewModuleFromFile("local.wasm")
		if err != nil {
			panic(err)
		}

		fn, err := module.StartFunction()
		if err != nil {
			panic(err)
		}

		engine, err := wasp.NewEngine(
			wasp.WithExternalFunc(consoleLog),
		)
		if err != nil {
			panic(err)
		}

		if _, err := engine.Call(module, fn); err != nil {
			panic(err)
		}
	}

	//wat := `(module
	//  (import "console" "log" (func $log (param i32)))
	//  (func $main
	//
	//	(local $var i32) ;; create a local variable named $var
	//	(local.set $var (i32.const 10)) ;; set $var to 10
	//	local.get $var ;; load $var onto the stack
	//	call $log ;; log the result
	//
	//  )
	//  (start $main)
	//)`
	//wasmBytes, err := wasmtime.Wat2Wasm(wat)
	//if err != nil {
	//	panic(err)
	//}
	//
	//f, _ := os.ReadFile("local.wasm")
	//if len(f) != len(wasmBytes) {
	//	panic("wasm bytes length mismatch")
	//}
	//for i := range f {
	//	if f[i] != wasmBytes[i] {
	//		panic("wasm bytes content mismatch")
	//	}
	//}
	//
	//module, err = wasp.NewModule(wasmBytes)
	//if err != nil {
	//	panic(err)
	//}
}
