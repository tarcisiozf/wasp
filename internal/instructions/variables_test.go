package instructions_test

import (
	"github.com/tarcisiozf/wasp"
	"github.com/tarcisiozf/wasp/tests"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocals(t *testing.T) {
	linker := wasp.NewLinker()
	logSpy := tests.NewSpy()
	assert.NoError(t, linker.Define("console", "log", logSpy.Called))

	testEnv := tests.NewEnvironment(
		tests.WithInstanceOptions(
			wasp.WithLinker(linker),
		),
	)
	build, err := testEnv.BuildWat(t, `
		(module
		  (import "console" "log" (func $log (param i32)))
		  (func $main
		
			(local $var i32) ;; create a local variable named $var
			(local.set $var (i32.const 10)) ;; set $var to 10
			local.get $var ;; load $var onto the stack
			call $log ;; log the result
		
		  )
		  (start $main)
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

	logSpy.CalledOnce(t)
	logSpy.FirstCall().CalledWith(t, 10)
}

func TestGlobals(t *testing.T) {
	linker := wasp.NewLinker()
	logSpy := tests.NewSpy()
	assert.NoError(t, linker.Define("console", "log", logSpy.Called))

	testEnv := tests.NewEnvironment(
		tests.WithInstanceOptions(
			wasp.WithLinker(linker),
		),
	)
	build, err := testEnv.BuildWat(t, `
		(module
		  (import "console" "log" (func $log (param i32)))
		  (global $var (mut i32) (i32.const 0))
		  (func $main
			i32.const 10 ;; load a number onto the stack
			global.set $var ;; set the $var
		
			global.get $var ;; load $var onto the stack
			call $log ;; log the result
		  )
		  (start $main)
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

	logSpy.CalledOnce(t)
	logSpy.FirstCall().CalledWith(t, 10)
}
