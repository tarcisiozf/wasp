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
	if len(args) < 1 {
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

		store := wasp.NewStore(module)

		linker := wasp.NewLinker()

		dg(linker, store)

		sp := wasi.NewWasiSnapshotPreview1()
		sp.SetArgs([]string{"doom1.wad"}) // Pass remaining args to WASI
		sp.AddPreopen(3, ".")             // Preopen current directory as fd 3
		//sp.AddPreopen(4, "doom1.wad")
		sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use
		if err := sp.Register(linker); err != nil {
			println("Error defining function:", err.Error())
			os.Exit(1)
		}

		options := []wasp.InstanceOption{
			wasp.WithLinker(linker),
		}
		for _, arg := range args {
			switch arg {
			case "-v", "--verbose":
				options = append(options, wasp.Verbose())
			}
		}

		instance, err := wasp.NewInstance(
			module,
			store,
			options...,
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

func dg(linker *wasp.Linker, store *wasp.Store) {
	var started time.Time
	linker.Define("dg", "init", func() {
		fmt.Println("DG initialized")
		started = time.Now()
	})

	linker.Define("dg", "draw_frame", func(ptr, resx, resy int32) {
		fmt.Printf("DB draw_frame called with ptr=%d, resx=%d, resy=%d\n", ptr, resx, resy)
	})

	linker.Define("dg", "sleep_ms", func(ms int32) {
		fmt.Printf("DB sleep_ms called with ms=%d\n", ms)
	})

	linker.Define("dg", "get_ticks_ms", func() int32 {
		elapsed := time.Since(started)
		ms := int32(elapsed.Milliseconds())
		fmt.Printf("DB get_ticks_ms called, returning %d ms\n", ms)
		return ms
	})

	linker.Define("dg", "get_key", func() int32 {
		fmt.Println("DB get_key called")
		return -1
	})

	linker.Define("dg", "set_window_title", func(ptr, len int32) {
		fmt.Printf("DB set_window_title called with ptr=%d, len=%d\n", ptr, len)
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
