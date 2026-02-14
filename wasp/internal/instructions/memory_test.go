package instructions_test

import (
	"testing"
	"wasp/wasp"
	"wasp/wasp/tests"

	"github.com/stretchr/testify/assert"
)

func TestMemory(t *testing.T) {
	linker := wasp.NewLinker()
	logSpy := tests.NewSpy()
	assert.NoError(t, linker.Define("console", "log", func(arg any) {
		logSpy.Called(arg)
	}))

	testEnv := tests.NewEnvironment(
		tests.WithInstanceOptions(
			wasp.WithLinker(linker),
		),
	)
	build, err := tests.BuildWat(t, `
		(module
		  (import "console" "log" (func $log (param i32)))
		  (memory 1 2) ;; default memory with one page and max of 2 pages
		
		  (func $main
			;; get size
			memory.size
			call $log ;; log the result (1)
		
			;; grow default memory by 1 page
			i32.const 1
			memory.grow
		
			;;get size again
			memory.size
			call $log ;; log the result (2)

			drop
		  )
		  (start $main) ;; call immediately on loading
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	if _, err := instance.RunStart(); err != nil {
		t.Fatalf("failed to run start function: %v", err)
	}

	logSpy.CallCount(t, 2)
	logSpy.OnCall(0).CalledWith(t, 1)
	logSpy.OnCall(1).CalledWith(t, 2)
}

func TestMemoryLoad(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := tests.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "load_first_item_in_mem") (param $num i32) (result i32)
			i32.const 0 ;; offset in memory to store the value
			local.get $num
			i32.store ;; store the value at the first byte of memory
		
			i32.const 0 ;; offset in memory to load from
			;; load first item in memory and return the result
			i32.load
		  )
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("load_first_item_in_mem", 100)
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int32(100)}, results)
}
