package main

import (
	"os"
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

	instance, err := wasp.NewInstance(
		module,
		store,
		wasp.WithLinker(linker),
		wasp.Verbose(),
	)
	if err != nil {
		println("Error creating instance of module:", err.Error())
		os.Exit(1)
	}

	fn, err := module.GetExportedFunction("_start")
	if err != nil {
		println("Error getting start function:", err.Error())
		os.Exit(1)
	}

	_, _ = instance.Call(fn)
	_ = instance.Tick()
}
