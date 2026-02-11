package main

import (
	"os"
	"wasp/wasp"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		println("Usage: run <wasm file>")
		os.Exit(1)
	}

	module, err := wasp.NewModuleFromFile(args[0])
	if err != nil {
		println("Error loading module:", err.Error())
		os.Exit(1)
	}

	_ = module
}
