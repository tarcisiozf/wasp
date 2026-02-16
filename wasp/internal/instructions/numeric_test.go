package instructions_test

import (
	"testing"
	"wasp/wasp/tests"

	"github.com/stretchr/testify/assert"
)

func TestConst(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i64 f32 f64)
			i32.const 10
			i64.const 20
			f32.const 30.0
			f64.const 40.0
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Len(t, results, 4)
	assert.Equal(t,
		[]any{int32(10), int64(20), float32(30.0), float64(40.0)},
		results,
	)
}

func TestF32ConvertI32S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result f32 f32 f32)
			i32.const 42
			f32.convert_i32_s
			i32.const 100
			f32.convert_i32_s
			i32.const 0
			f32.convert_i32_s
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Len(t, results, 3)
	assert.Equal(t,
		[]any{float32(42.0), float32(100.0), float32(0.0)},
		results,
	)
}

func TestF32ReinterpretI32(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result f32 f32 f32)
			i32.const 0
			f32.reinterpret_i32
			i32.const 1065353216
			f32.reinterpret_i32
			i32.const 1078530011
			f32.reinterpret_i32
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	// 0 -> 0.0, 1065353216 -> 1.0, 1078530011 -> 3.14159...
	assert.Len(t, results, 3)
	assert.Equal(t, float32(0.0), results[0])
	assert.Equal(t, float32(1.0), results[1])
	assert.InDelta(t, float32(3.14159), results[2], 0.0001)
}

func TestF64ConvertI32S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result f64 f64 f64)
			i32.const 42
			f64.convert_i32_s
			i32.const 100
			f64.convert_i32_s
			i32.const 0
			f64.convert_i32_s
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Len(t, results, 3)
	assert.Equal(t,
		[]any{float64(42.0), float64(100.0), float64(0.0)},
		results,
	)
}

func TestF64Copysign(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result f64 f64 f64)
			f64.const 5.0
			f64.const -1.0
			f64.copysign
			f64.const -5.0
			f64.const 1.0
			f64.copysign
			f64.const 3.14
			f64.const -0.0
			f64.copysign
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	// copysign(5.0, -1.0) = -5.0, copysign(-5.0, 1.0) = 5.0, copysign(3.14, -0.0) = -3.14
	assert.Len(t, results, 3)
	assert.Equal(t, float64(-5.0), results[0])
	assert.Equal(t, float64(5.0), results[1])
	assert.Equal(t, float64(-3.14), results[2])
}

func TestF64PromoteF32(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result f64 f64 f64)
			f32.const 3.14
			f64.promote_f32
			f32.const 0.0
			f64.promote_f32
			f32.const -1.5
			f64.promote_f32
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	assert.Len(t, results, 3)
	assert.InDelta(t, float64(3.14), results[0], 0.0001)
	assert.Equal(t, float64(0.0), results[1])
	assert.Equal(t, float64(-1.5), results[2])
}

func TestF64ReinterpretI64(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result f64 f64 f64)
			i64.const 0
			f64.reinterpret_i64
			i64.const 4607182418800017408
			f64.reinterpret_i64
			i64.const 4614256656552045848
			f64.reinterpret_i64
		  )
		  (export "subject" (func $subject))
		)
	`)
	if err != nil {
		t.Fatalf("failed to build wat: %v", err)
	}

	instance, err := testEnv.CreateInstance(build.Wasm)
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	_, results, err := instance.RunExport("subject")
	if err != nil {
		t.Fatalf("failed to run function: %v", err)
	}

	// 0 -> 0.0, 4607182418800017408 -> 1.0, 4614256656552045848 -> 3.14159...
	assert.Len(t, results, 3)
	assert.Equal(t, float64(0.0), results[0])
	assert.Equal(t, float64(1.0), results[1])
	assert.InDelta(t, float64(3.14159), results[2], 0.0001)
}
