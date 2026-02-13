package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"wasp/wasp"
)

var linker = wasp.NewLinker()

var examples = map[string]string{
	"eq": `(module
	  (import "env" "log_bool" (func $log_bool (param i32)))
	  (func $main
		;; load 10 and 2 onto the stack
		i32.const 10
		i32.const 2
	
		i32.eq ;; check if 10 is equal to 2
		call $log_bool ;; log the result
	  )
	  (start $main)
	)`,
}

func main() {
	if err := linker.Define("console", "log", log); err != nil {
		panic(err)
	}

	if err := linker.Define("env", "log_bool", log); err != nil {
		panic(err)
	}

	for name, wat := range examples {
		watPath := fmt.Sprintf("builds/%s.wat", name)
		outputPath := fmt.Sprintf("builds/%s.wasm", name)
		bytecodePath := fmt.Sprintf("builds/%s.s", name)

		_, err := os.Stat(watPath)
		if os.IsNotExist(err) {
			if err := os.WriteFile(watPath, []byte(wat), 0644); err != nil {
				panic(fmt.Sprintf("failed to write wat file: %v", err))
			}

			cmd := exec.Command("wat2wasm", watPath, "-o", outputPath, "-v")
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				os.Remove(watPath)
				fmt.Fprintf(os.Stderr, "failed to run wat2wasm: %v\nstdout: %s\nstderr: %s\n", err, stdout.String(), stderr.String())
			}

			if err := os.WriteFile(bytecodePath, stderr.Bytes(), 0644); err != nil {
				panic(fmt.Sprintf("failed to write bytecode file: %v", err))
			}
		}

		results, err := run(outputPath)
		if err != nil {
			panic(fmt.Sprintf("failed to run wasm module %s: %v", name, err))
		}

		fmt.Printf("results of %s: %v\n", name, results)
	}
}

func run(path string) ([]any, error) {
	module, err := wasp.NewModuleFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load module: %w", err)
	}

	fn, err := module.StartFunction()
	if err != nil {
		return nil, fmt.Errorf("failed to get start function: %w", err)
	}

	runtime, err := wasp.NewInstance(
		module,
		wasp.WithLinker(linker),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	results, err := runtime.Call(fn)
	if err != nil {
		return nil, fmt.Errorf("failed to execute start function: %w", err)
	}

	return results, nil
}

func log(args ...any) {
	fmt.Println(args...)
}
