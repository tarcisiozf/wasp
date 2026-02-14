package instructions_test

import (
	"fmt"
	"testing"
	"wasp/wasp/tests"

	"github.com/stretchr/testify/assert"
)

func TestBranchBlock(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := tests.BuildWat(t, `
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
	build, err := tests.BuildWat(t, `
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
	build, err := tests.BuildWat(t, `
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
	build, err := tests.BuildWat(t, `
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
	build, err := tests.BuildWat(t, `
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
	build, err := tests.BuildWat(t, `
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
	build, err := tests.BuildWat(t, `
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
