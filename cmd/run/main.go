package main

import (
	"fmt"
	"os"
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

	linker := wasp.NewLinker()
	if requiresWasi {
		sp := wasi.NewWasiSnapshotPreview1()
		if err := sp.Register(linker); err != nil {
			println("Error registering WASI snapshot preview 1:", err.Error())
			os.Exit(1)
		}
	}

	var autoRun = true
	options := []wasp.InstanceOption{
		wasp.WithLinker(linker),
	}
	for _, arg := range args {
		switch arg {
		case "--verbose", "-v":
			options = append(options, wasp.Verbose())
		case "--dry-run":
			autoRun = false
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
		fn, err := findCandidateForStartFunc(module)
		if err != nil {
			println("Error getting start function:", err.Error())
			os.Exit(1)
		}

		if _, err = instance.Call(fn); err != nil {
			println("Error calling start function:", err.Error())
			os.Exit(1)
		}

		if err := instance.Tick(); err != nil {
			println("Error ticking instance:", err.Error())
			os.Exit(1)
		}
	}
}

func findCandidateForStartFunc(module *wasp.Module) (int, error) {
	if fn, err := module.GetStartFunction(); err == nil {
		return fn, nil
	}

	if fn, err := module.GetExportedFunction("_start"); err == nil {
		return fn, nil
	}

	return -1, fmt.Errorf("could not find candidate for start function")
}
