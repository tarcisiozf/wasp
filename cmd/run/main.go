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
		_ = module
		fmt.Println("WASM loaded in ", time.Since(start))
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
