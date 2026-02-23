package main

import (
	"fmt"
	"os"

	"github.com/tarcisiozf/wasp"
	"github.com/tarcisiozf/wasp/wasi"
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

	module, err := wasp.NewModule(wasm)
	if err != nil {
		println("Error loading module:", err.Error())
		os.Exit(1)
	}

	var requiresWasi bool
	for _, imp := range module.Imports() {
		if imp.ModuleName == "wasi_snapshot_preview1" {
			requiresWasi = true
			break
		}
	}

	store := wasp.NewStore(module)
	defaultMem := store.Memories[0]
	defaultMem.Grow(200)

	linker := wasp.NewLinker()

	if requiresWasi {
		sp := wasi.NewWasiSnapshotPreview1()
		sp.SetArgs(args)                // Pass remaining args to WASI
		sp.AddPreopen(3, ".")           // Preopen current directory as fd 3
		sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use
		if err := sp.Register(linker); err != nil {
			println("Error registering WASI snapshot preview 1:", err.Error())
			os.Exit(1)
		}
	}

	var autoRun = true
	var runFunc string
	options := []wasp.InstanceOption{
		wasp.WithLinker(linker),
	}
	for i, arg := range args {
		switch arg {
		case "--verbose", "-v":
			options = append(options, wasp.Verbose())
		case "--dry-run":
			autoRun = false
		case "--func", "-f":
			runFunc = args[i+1]
		}
	}

	instance, err := wasp.NewInstance(
		module,
		store,
		options...,
	)
	if err != nil {
		println("Error creating instance of module:", err.Error())
		os.Exit(1)
	}

	if autoRun {
		funcref, err := findCandidateForStartFunc(module, runFunc)
		if err != nil {
			println("Error getting start function:", err.Error())
			os.Exit(1)
		}

		if _, err = instance.Call(funcref); err != nil {
			println("Error calling start function:", err.Error())
			os.Exit(1)
		}

		if err := instance.Run(); err != nil {
			println("Error running instance:", err.Error())
			os.Exit(1)
		}
	}
}

func findCandidateForStartFunc(module *wasp.Module, target string) (int, error) {
	if target != "" {
		return module.GetExportedFunction(target)
	}

	if fn, err := module.GetStartFunction(); err == nil {
		return fn, nil
	}

	if fn, err := module.GetExportedFunction("_start"); err == nil {
		return fn, nil
	}

	return -1, fmt.Errorf("could not find candidate for start function")
}
