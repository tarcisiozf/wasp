package tests

import (
	"bytes"
	"fmt"
	"github.com/tarcisiozf/wasp/debug"
	"os"
	"os/exec"
	"strings"

	"github.com/bytecodealliance/wasmtime-go"
)

type Build struct {
	Wat  string
	Asm  string
	Wasm []byte
}

type Wat2WasmBuilder struct{}

func (b *Wat2WasmBuilder) Build(dir, wat string) (Build, error) {
	watPath := dir + "/tmp.wat"
	outputPath := dir + "/tmp.wasm"

	if err := os.WriteFile(watPath, []byte(wat), 0644); err != nil {
		return Build{}, fmt.Errorf("failed to write wat: %v", err)
	}

	cmd := exec.Command("wat2wasm", watPath, "-o", outputPath, "-v")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Build{}, fmt.Errorf("failed to run wat2wasm: %v\nstdout: %s\nstderr: %s\n", err, stdout.String(), stderr.String())
	}
	asm := stderr.String() // wat2wasm outputs the disassembly to stderr

	wasm, err := os.ReadFile(outputPath)
	if err != nil {
		return Build{}, fmt.Errorf("failed to read wasm file: %v", err)
	}

	return Build{
		Wat:  wat,
		Asm:  asm,
		Wasm: wasm,
	}, nil
}

type WasmtimeWatBuilder struct{}

func (b *WasmtimeWatBuilder) Build(_, wat string) (Build, error) {
	wasm, err := wasmtime.Wat2Wasm(wat)
	if err != nil {
		return Build{}, fmt.Errorf("failed to convert wat to wasm: %v", err)
	}
	var sb strings.Builder
	if err := debug.WasmToString(&sb, wasm); err != nil {
		fmt.Println(sb.String())
		return Build{}, fmt.Errorf("failed to disassemble wasm: %v", err)
	}
	return Build{
		Wat:  wat,
		Asm:  sb.String(),
		Wasm: wasm,
	}, nil
}
