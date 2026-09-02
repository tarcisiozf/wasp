package instructions_test

import (
	"fmt"
	"testing"

	"github.com/tarcisiozf/wasp"
	"github.com/tarcisiozf/wasp/tests"

	"github.com/stretchr/testify/assert"
)

func TestBranchBlock(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject
			(block $my_block
		
			  ;; Break out of the block
			  ;; If this is removed, the code will throw an error when it reaches unreachable
			  br $my_block
		
			  ;; The code will never reach this point since we broke out of the block
			  unreachable
		
			)
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, _, err = instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}
}

func TestBranchNestedBlocks(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32)
			(block $outer (result i32)
			  (block $inner (result i32)
				;; Push a value and branch to outer block
				i32.const 42
				br $outer
				;; This should never be reached
				i32.const 99
			  )
			)
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(42)}, results)
}

func TestBranchLoop(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32)
			(local $counter i32)
			;; Initialize counter to 0
			i32.const 0
			local.set $counter
			
			(block $exit (result i32)
			  (loop $loop
				;; Increment counter
				local.get $counter
				i32.const 1
				i32.add
				local.set $counter
				
				;; If counter == 5, exit with counter value
				local.get $counter
				i32.const 5
				i32.eq
				(if
				  (then
					local.get $counter
					br $exit
				  )
				)
				
				;; Otherwise, continue loop
				br $loop
			  )
			  ;; Default value (should not be reached)
			  i32.const 0
			)
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(5)}, results)
}

func TestBrIf(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32)
			(block $exit (result i32)
			  ;; Push value 10 for the block result
			  i32.const 10
			  ;; Condition is true (1), so we branch
			  i32.const 1
			  br_if $exit
			  ;; This should not be reached
			  drop
			  i32.const 99
			)
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(10)}, results)
}

func TestBrIfFalse(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32)
			(block $exit (result i32)
			  ;; Push value 10
			  i32.const 10
			  ;; Condition is false (0), so we don't branch
			  i32.const 0
			  br_if $exit
			  ;; br_if didn't branch, so drop the 10 and push 99
			  drop
			  i32.const 99
			)
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(99)}, results)
}

func TestBranchIf(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32)
			(block $outer (result i32)
			  i32.const 1
			  (if (result i32)
				(then
				  ;; Branch out of the if with value 42
				  i32.const 42
				  br $outer
				)
				(else
				  i32.const 0
				)
			  )
			)
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(42)}, results)
}

func TestBranchLoopContinue(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32)
			(local $i i32)
			(local $sum i32)
			;; sum = 0, i = 0
			i32.const 0
			local.set $sum
			i32.const 0
			local.set $i
			
			(loop $loop
			  ;; i = i + 1
			  local.get $i
			  i32.const 1
			  i32.add
			  local.set $i
			  
			  ;; sum = sum + i
			  local.get $sum
			  local.get $i
			  i32.add
			  local.set $sum
			  
			  ;; if i != 5, continue loop (i.e., if i == 5, don't branch)
			  local.get $i
			  i32.const 5
			  i32.eq
			  (if
			    (then)
			    (else
			      br $loop
			    )
			  )
			)
			
			;; return sum (should be 1+2+3+4+5 = 15)
			local.get $sum
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(15)}, results)
}

func TestCall(t *testing.T) {
	linker := wasp.NewLinker()
	err := linker.Define("env", "number", func() int32 {
		return 42
	})
	if err != nil {
		t.Fatalf("failed to define host function: %v", err)
	}

	testEnv := tests.NewEnvironment(
		tests.WithInstanceOptions(
			wasp.WithLinker(linker),
		),
	)
	build, err := testEnv.BuildWat(t, `
		(module
		  (import "env" "number" (func $give_number (result i32)))
		
		  (func (export "subject") (result i32)
			call $give_number
		  )
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(42)}, results)
}

func TestCallIndirect(t *testing.T) {
	testEnv := tests.NewEnvironment(
		tests.WithWasmtimeBuilder(),
	)
	build, err := testEnv.BuildWat(t, `
		(module
		  ;; type 0: (i32, i32) -> i32
		  (type (func (param i32 i32) (result i32)))
		
		  ;; table 0 with 2 function refs
		  (table 2 funcref)
		
		  ;; func 0
		  (func (param i32 i32) (result i32)
			local.get 0
			local.get 1
			i32.add
		  )
		
		  ;; func 1
		  (func (param i32 i32) (result i32)
			local.get 0
			local.get 1
			i32.mul
		  )
		
		  ;; initialize table[0]=func0, table[1]=func1
		  (elem (i32.const 0) 0 1)
		
		  ;; dispatch(a, b, op) -> i32
		  (func (export "dispatch") (param i32 i32 i32) (result i32)
			;; push args first
			local.get 0
			local.get 1
			;; then the table index
			local.get 2
			;; indirect call expects the type index (here: 0)
			call_indirect (type 0)
		  )
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("dispatch", 20, 22, 0) // should call func0 (add)
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(42)}, results)

	// Test calling multiply function (index 1)
	_, results, err = instance.RunExport("dispatch", 6, 7, 1) // should call func1 (mul)
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(42)}, results)
}

func TestCallIndirectTypeMismatch(t *testing.T) {
	testEnv := tests.NewEnvironment(
		tests.WithWasmtimeBuilder(),
	)
	build, err := testEnv.BuildWat(t, `
		(module
		  ;; type 0: (i32, i32) -> i32
		  (type (func (param i32 i32) (result i32)))
		  ;; type 1: (i32) -> i32
		  (type (func (param i32) (result i32)))
		
		  ;; table 0 with 2 function refs
		  (table 2 funcref)
		
		  ;; func 0 has type 0: (i32, i32) -> i32
		  (func (param i32 i32) (result i32)
			local.get 0
			local.get 1
			i32.add
		  )
		
		  ;; func 1 has type 1: (i32) -> i32
		  (func (param i32) (result i32)
			local.get 0
			i32.const 2
			i32.mul
		  )
		
		  ;; initialize table[0]=func0, table[1]=func1
		  (elem (i32.const 0) 0 1)
		
		  ;; This function tries to call func1 (type 1) but uses type 0
		  (func (export "bad_call") (param i32 i32) (result i32)
			local.get 0
			local.get 1
			;; call table[1] which is func1 (type 1), but expect type 0
			i32.const 1
			call_indirect (type 0)
		  )
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	// This should fail due to type mismatch
	_, _, err = instance.RunExport("bad_call", 10, 20)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type mismatch")
}

func TestReturnCall(t *testing.T) {
	testEnv := tests.NewEnvironment(
		tests.WithWasmtimeBuilder(),
	)
	build, err := testEnv.BuildWat(t, `
		(module
		  ;; Calculate the factorial of a number
		  (func $fac (export "fac") (param $x i64) (result i64)
			;; Call the fac-aux function with $x and 1 parameters
			(return_call $fac-aux (local.get $x) (i64.const 1))
		  )

		  ;; Perform the factorial calculation
		  (func $fac-aux (param $x i64) (param $r i64) (result i64)
			;; If $x is zero, return the accumulated result $r
			(if (result i64) (i64.eqz (local.get $x))
			  (then (return (local.get $r)))
			  (else
				;; Otherwise, recursively call fac-aux with $x-1 and $x*$r
				(return_call $fac-aux
				  (i64.sub (local.get $x) (i64.const 1))
				  (i64.mul (local.get $x) (local.get $r))
				)
			  )
			)
		  )
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("fac", int64(5))
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int64(120)}, results)
}

func TestReturn(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func (export "subject") (result i32)
			;; load 10 onto the stack
			i32.const 10
			;; load 90 onto the stack
			i32.const 90
			;; return the second value (90); the first is discarded
			return
		  )
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(90)}, results)
}

func TestSelect(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func (export "subject") (result i32)
			;; load two values onto the stack
			i32.const 10
			i32.const 20
		
			;; change to 1 (true) to get the first value (10)
			i32.const 0
			select
		  )
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Equal(t, []any{int32(20)}, results)
}

func TestBrTable(t *testing.T) {
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
		  ;; Import the browser console object, which you'll need to pass in from JavaScript
		  (import "console" "log" (func $log (param i32)))
		
		  (func
			;; Label each block for easy reference
			;; (they can also be referenced by their index)
			(block $outer_block
			  (block $middle_block
				(block $inner_block
		
				  ;; Choose which block to break out of based on their order in the br_table
				  ;; 0 is $inner_block, 1 is $outer_block, 2 is $middle_block
				  i32.const 0
		
				  ;; Create a br_table with three targets
				  (br_table $inner_block $outer_block $middle_block)
		
				  ;; The code will never reach this point since we broke out of the block
				  unreachable
		
				)
		
				;; If you jump out of $inner_block but stay in $middle_block,
				;; 42 will be logged
				;; If you jump out of $middle_block also,
				;; by jumping out of either $middle_block or $outer_block,
				;; this will be skipped
				i32.const 42
				call $log
			  )
			)
		  )

		  (start 1)
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	fmt.Println(build.Asm)

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, err = instance.RunStart()
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	logSpy.CalledOnce(t)
	logSpy.FirstCall().CalledWith(t, 42)
}
