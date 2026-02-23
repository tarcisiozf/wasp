package main

import (
	"github.com/tarcisiozf/wasp/debug"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		println("Usage: asm <wasm file>")
		os.Exit(1)
	}

	filename := args[0]
	wasm, err := os.ReadFile(filename)
	if err != nil {
		println("Error reading file:", err.Error())
		os.Exit(1)
	}

	err = debug.WasmToString(os.Stdout, wasm)
	if err != nil {
		println("Error converting WASM to string:", err.Error())
		os.Exit(1)
	}
}
