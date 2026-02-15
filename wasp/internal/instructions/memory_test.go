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
	build, err := testEnv.BuildWat(t, `
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
	build, err := testEnv.BuildWat(t, `
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

func TestI64Load(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i64)
			;; Store value 100 at offset 0
			i32.const 0
			i32.const 100
			i32.store
			;; Store value 200 at offset 4
			i32.const 4
			i32.const 200
			i32.store
		
			i32.const 0
			i64.load
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	// Little-endian: low 4 bytes = 100 (0x64), high 4 bytes = 200 (0xC8)
	// Expected: 200 << 32 | 100 = 858993459300
	expected := int64(200)<<32 | int64(100)
	assert.Equal(t, []any{expected}, results)
}

func TestF32Load(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result f32)
			i32.const 0
			f32.const 3.14
			f32.store
		
			i32.const 0
			f32.load
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.InDelta(t, float32(3.14), results[0].(float32), 0.001)
}

func TestF64Load(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result f64)
			i32.const 0
			f64.const 3.141592653589793
			f64.store
		
			i32.const 0
			f64.load
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.InDelta(t, float64(3.141592653589793), results[0].(float64), 0.0000001)
}

func TestI32Load8S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i32)
			i32.const 0
			i32.const 0xFF ;; -1 as signed byte
			i32.store8
		
			i32.const 0
			i32.load8_s ;; sign-extend to i32
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int32(-1)}, results)
}

func TestI32Load8U(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i32)
			i32.const 0
			i32.const 0xFF
			i32.store8
		
			i32.const 0
			i32.load8_u ;; zero-extend to i32
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int32(255)}, results)
}

func TestI32Load16S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i32)
			i32.const 0
			i32.const 0xFFFF ;; -1 as signed 16-bit
			i32.store16
		
			i32.const 0
			i32.load16_s ;; sign-extend to i32
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int32(-1)}, results)
}

func TestI32Load16U(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i32)
			i32.const 0
			i32.const 0xFFFF
			i32.store16
		
			i32.const 0
			i32.load16_u ;; zero-extend to i32
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int32(65535)}, results)
}

func TestI64Load8S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i64)
			i32.const 0
			i32.const 0x80 ;; -128 as signed byte
			i32.store8
		
			i32.const 0
			i64.load8_s ;; sign-extend to i64
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int64(-128)}, results)
}

func TestI64Load8U(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i64)
			i32.const 0
			i32.const 0x80
			i32.store8
		
			i32.const 0
			i64.load8_u ;; zero-extend to i64
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int64(128)}, results)
}

func TestI64Load16S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i64)
			i32.const 0
			i32.const 0x8000 ;; -32768 as signed 16-bit
			i32.store16
		
			i32.const 0
			i64.load16_s ;; sign-extend to i64
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int64(-32768)}, results)
}

func TestI64Load16U(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i64)
			i32.const 0
			i32.const 0x8000
			i32.store16
		
			i32.const 0
			i64.load16_u ;; zero-extend to i64
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int64(32768)}, results)
}

func TestI64Load32S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i64)
			i32.const 0
			i32.const 0x80000000 ;; -2147483648 as signed 32-bit
			i32.store
		
			i32.const 0
			i64.load32_s ;; sign-extend to i64
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int64(-2147483648)}, results)
}

func TestI64Load32U(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i64)
			i32.const 0
			i32.const 0x80000000
			i32.store
		
			i32.const 0
			i64.load32_u ;; zero-extend to i64
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int64(2147483648)}, results)
}

func TestMemoryLoadWithOffset(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (memory $memory 1)
		  (export "memory" (memory $memory))
		
		  (func (export "test") (result i32)
			;; Store value 42 at address 8
			i32.const 8
			i32.const 42
			i32.store
		
			;; Load from address 4 with offset 4 (effective address = 4 + 4 = 8)
			i32.const 4
			i32.load offset=4
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

	_, results, err := instance.RunExport("test")
	if err != nil {
		t.Fatalf("failed to run export function: %v", err)
	}

	assert.Equal(t, []any{int32(42)}, results)
}
