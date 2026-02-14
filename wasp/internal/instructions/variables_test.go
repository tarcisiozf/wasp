package instructions_test

import (
	"testing"
	"wasp/wasp"
	"wasp/wasp/tests"

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
	_, _, err := testEnv.RunWat(t, `
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
	assert.NoError(t, err)

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
	_, _, err := testEnv.RunWat(t, `
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
	assert.NoError(t, err)

	logSpy.CalledOnce(t)
	logSpy.FirstCall().CalledWith(t, 10)
}
