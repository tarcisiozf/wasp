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

	"if": `(module
	  ;; import the browser console object, you'll need to pass this in from JavaScript
	  (import "console" "log" (func $log (param i32)))
	
	  (func
		i32.const 0 ;; change to positive number (true) if you want to run the if block
		(if
		  (then
			i32.const 1
			call $log ;; should log '1'
		  )
		  (else
			i32.const 0
			call $log ;; should log '0'
		  )
		)
	  )
	
	  (start 1) ;; run the first function automatically
	)`,

	"block": `
		(module
		  ;; import the browser console object, you'll need to pass this in from JavaScript
		  (import "console" "log" (func $log (param i32)))
		
		  ;; create a function that takes in a number as a param,
		  ;; and logs that number if it's not equal to 100.
		  (func $main
			(block $my_block

				i32.const 100
			  i32.const 100
			  i32.eq
		
			  (if
				(then
		
				  ;; branch to the end of the block
				  br $my_block
		
				)
			  )
		
				unreachable
			  ;; not reachable when $num is 100
			  ;; local.get $num
			  ;; call $log
		
			)
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

		fmt.Printf("### running %s.wasm ###\n", name)

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
