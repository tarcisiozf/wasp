package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"wasp/wasi"
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

		dg(linker)

		sp := wasi.NewWasiSnapshotPreview1()
		sp.SetArgs(os.Args[1:]) // Pass remaining args to WASI
		sp.AddPreopen(3, ".")   // Preopen current directory as fd 3
		if err := sp.Register(linker); err != nil {
			println("Error defining function:", err.Error())
			os.Exit(1)
		}

		store := wasp.NewStore(module)

		instance, err := wasp.NewInstance(
			module,
			store,
			wasp.WithLinker(linker),
			wasp.Verbose(),
		)
		if err != nil {
			println("Error creating runtime:", err.Error())
			os.Exit(1)
		}

		fn, err := module.GetExportedFunction("_start")
		if err != nil {
			println("Error getting function:", err.Error())
			os.Exit(1)
		}

		var results []any
		elapsed(func() {
			cf, err := instance.Call(fn)
			if err != nil {
				println("Error calling function:", err.Error())
				os.Exit(1)
			}
			if err := instance.Tick(); err != nil {
				println("Error during execution:", err.Error())
				os.Exit(1)
			}
			results, err = cf.Results()
			if err != nil {
				println("Error getting results:", err.Error())
				os.Exit(1)
			}
		})

		fmt.Printf("Results: %v\n", results)
	})
}

func dg(linker *wasp.Linker) {
	linker.Define("dg", "init", func() {

	})

	linker.Define("dg", "draw_frame", func(ptr, resx, resy int32) {

	})

	linker.Define("dg", "sleep_ms", func(ms int32) {

	})

	linker.Define("dg", "get_ticks_ms", func() int32 {
		return 0
	})

	linker.Define("dg", "get_key", func() int32 {
		return -1
	})

	linker.Define("dg", "set_window_title", func(ptr int32) {
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
