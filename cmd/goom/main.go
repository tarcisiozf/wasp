package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"wasp/wasp"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		println("Usage: run <wasm file>")
		os.Exit(1)
	}

	wasm, err := os.ReadFile(args[0])
	if err != nil {
		println("Error reading file:", err.Error())
		os.Exit(1)
	}

	memstats(func() {
		start := time.Now()
		module, err := wasp.NewModule(wasm)
		if err != nil {
			println("Error loading module:", err.Error())
			os.Exit(1)
		}
		fmt.Println("WASM loaded in ", time.Since(start))

		linker := wasp.NewLinker()

		r, err := wasp.NewRuntime(
			module,
			wasp.WithLinker(linker),
		)
		if err != nil {
			println("Error creating runtime:", err.Error())
			os.Exit(1)
		}

		funcs := []string{
			"emscripten_stack_init",
			"__wasm_call_ctors",
			//"__main_argc_argv",
		}
		for _, name := range funcs {
			fmt.Println("Calling function:", name)

			fn, err := module.GetExportedFunction(name)
			if err != nil {
				println("Error getting function:", err.Error())
				os.Exit(1)
			}

			var results []any
			elapsed(func() {
				results, err = r.Call(fn)
				if err != nil {
					println("Error calling function:", err.Error())
					os.Exit(1)
				}
			})

			fmt.Printf("Results: %v\n", results)
		}
	})
}

func memstats(f func()) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)

	fmt.Printf("Memory usage: %d bytes\n", after.Alloc-before.Alloc)
	fmt.Printf("Heap allocations: %d bytes\n", after.HeapAlloc-before.HeapAlloc)
	fmt.Printf("Heap objects: %d\n", after.HeapObjects-before.HeapObjects)
}

func elapsed(f func()) {
	start := time.Now()
	f()
	fmt.Printf("Elapsed time: %s\n", time.Since(start))
}
