package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

type Build struct {
	Wat  string
	Asm  string
	Wasm []byte
}

func BuildWat(t *testing.T, wat string) (Build, error) {
	tmpDir := t.TempDir()
	watPath := tmpDir + "/tmp.wat"
	outputPath := tmpDir + "/tmp.wasm"

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
