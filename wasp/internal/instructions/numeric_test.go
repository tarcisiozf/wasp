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

func TestI32Clz(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32)
			i32.const 1
			i32.clz
			i32.const 0x80000000
			i32.clz
			i32.const 0
			i32.clz
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
	// clz(1) = 31, clz(0x80000000) = 0, clz(0) = 32
	assert.Len(t, results, 3)
	assert.Equal(t, int32(31), results[0])
	assert.Equal(t, int32(0), results[1])
	assert.Equal(t, int32(32), results[2])
}

func TestI32Ctz(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32)
			i32.const 1
			i32.ctz
			i32.const 0x80000000
			i32.ctz
			i32.const 0
			i32.ctz
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
	// ctz(1) = 0, ctz(0x80000000) = 31, ctz(0) = 32
	assert.Len(t, results, 3)
	assert.Equal(t, int32(0), results[0])
	assert.Equal(t, int32(31), results[1])
	assert.Equal(t, int32(32), results[2])
}

func TestI32DivU(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32)
			i32.const 10
			i32.const 3
			i32.div_u
			i32.const 100
			i32.const 7
			i32.div_u
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
	// 10/3 = 3, 100/7 = 14
	assert.Len(t, results, 2)
	assert.Equal(t, int32(3), results[0])
	assert.Equal(t, int32(14), results[1])
}

func TestI32RemSU(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32)
			i32.const 10
			i32.const 3
			i32.rem_s
			i32.const 10
			i32.const 3
			i32.rem_u
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
	// rem_s(10,3)=1, rem_u(10,3)=1
	assert.Len(t, results, 2)
	assert.Equal(t, int32(1), results[0])
	assert.Equal(t, int32(1), results[1])
}

func TestI32Comparisons(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32 i32 i32 i32 i32 i32)
			i32.const 1
			i32.const 5
			i32.lt_s
			i32.const 1
			i32.const 5
			i32.lt_u
			i32.const 5
			i32.const 1
			i32.gt_s
			i32.const 5
			i32.const 1
			i32.gt_u
			i32.const 5
			i32.const 5
			i32.le_s
			i32.const 5
			i32.const 5
			i32.le_u
			i32.const 5
			i32.const 5
			i32.ge_s
			i32.const 5
			i32.const 5
			i32.ge_u
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
	// lt_s(1,5)=1, lt_u(1,5)=1
	// gt_s(5,1)=1, gt_u(5,1)=1
	// le_s(5,5)=1, le_u(5,5)=1
	// ge_s(5,5)=1, ge_u(5,5)=1
	assert.Len(t, results, 8)
	assert.Equal(t, 1, results[0]) // lt_s
	assert.Equal(t, 1, results[1]) // lt_u
	assert.Equal(t, 1, results[2]) // gt_s
	assert.Equal(t, 1, results[3]) // gt_u
	assert.Equal(t, 1, results[4]) // le_s
	assert.Equal(t, 1, results[5]) // le_u
	assert.Equal(t, 1, results[6]) // ge_s
	assert.Equal(t, 1, results[7]) // ge_u
}

func TestI32ShiftRotate(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32 i32)
			i32.const 1
			i32.const 4
			i32.shl
			i32.const 16
			i32.const 2
			i32.shr_s
			i32.const 16
			i32.const 2
			i32.shr_u
			i32.const 1
			i32.const 33
			i32.shl
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
	// shl(1,4)=16, shr_s(16,2)=4, shr_u(16,2)=4
	// shl(1,33)=shl(1,1)=2 (mod 32)
	assert.Len(t, results, 4)
	assert.Equal(t, int32(16), results[0])
	assert.Equal(t, int32(4), results[1])
	assert.Equal(t, int32(4), results[2])
	assert.Equal(t, int32(2), results[3])
}

func TestI32Rotl(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32)
			i32.const 1
			i32.const 1
			i32.rotl
			i32.const 3
			i32.const 2
			i32.rotl
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
	// rotl(1,1)=2, rotl(3,2)=12
	assert.Len(t, results, 2)
	assert.Equal(t, int32(2), results[0])
	assert.Equal(t, int32(12), results[1])
}

func TestI64Clz(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64 i64)
			i64.const 1
			i64.clz
			i64.const 0x8000000000000000
			i64.clz
			i64.const 0
			i64.clz
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
	// clz(1) = 63, clz(0x8000000000000000) = 0, clz(0) = 64
	assert.Len(t, results, 3)
	assert.Equal(t, int64(63), results[0])
	assert.Equal(t, int64(0), results[1])
	assert.Equal(t, int64(64), results[2])
}

func TestI64Ctz(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64 i64)
			i64.const 1
			i64.ctz
			i64.const 0x8000000000000000
			i64.ctz
			i64.const 0
			i64.ctz
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
	// ctz(1) = 0, ctz(0x8000000000000000) = 63, ctz(0) = 64
	assert.Len(t, results, 3)
	assert.Equal(t, int64(0), results[0])
	assert.Equal(t, int64(63), results[1])
	assert.Equal(t, int64(64), results[2])
}

func TestI64DivU(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64)
			i64.const 10
			i64.const 3
			i64.div_u
			i64.const 100
			i64.const 7
			i64.div_u
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
	// 10/3 = 3, 100/7 = 14
	assert.Len(t, results, 2)
	assert.Equal(t, int64(3), results[0])
	assert.Equal(t, int64(14), results[1])
}

func TestI64RemSU(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64)
			i64.const 10
			i64.const 3
			i64.rem_s
			i64.const 10
			i64.const 3
			i64.rem_u
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
	// rem_s(10,3)=1, rem_u(10,3)=1
	assert.Len(t, results, 2)
	assert.Equal(t, int64(1), results[0])
	assert.Equal(t, int64(1), results[1])
}

func TestI64Comparisons(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32 i32 i32 i32 i32 i32)
			i64.const 1
			i64.const 5
			i64.lt_s
			i64.const 1
			i64.const 5
			i64.lt_u
			i64.const 5
			i64.const 1
			i64.gt_s
			i64.const 5
			i64.const 1
			i64.gt_u
			i64.const 5
			i64.const 5
			i64.le_s
			i64.const 5
			i64.const 5
			i64.le_u
			i64.const 5
			i64.const 5
			i64.ge_s
			i64.const 5
			i64.const 5
			i64.ge_u
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
	// lt_s(1,5)=1, lt_u(1,5)=1, gt_s(5,1)=1, gt_u(5,1)=1
	// le_s(5,5)=1, le_u(5,5)=1, ge_s(5,5)=1, ge_u(5,5)=1
	assert.Len(t, results, 8)
	assert.Equal(t, 1, results[0]) // lt_s
	assert.Equal(t, 1, results[1]) // lt_u
	assert.Equal(t, 1, results[2]) // gt_s
	assert.Equal(t, 1, results[3]) // gt_u
	assert.Equal(t, 1, results[4]) // le_s
	assert.Equal(t, 1, results[5]) // le_u
	assert.Equal(t, 1, results[6]) // ge_s
	assert.Equal(t, 1, results[7]) // ge_u
}

func TestI64ShiftRotate(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64 i64 i64 i64)
			i64.const 1
			i64.const 4
			i64.shl
			i64.const 16
			i64.const 2
			i64.shr_s
			i64.const 16
			i64.const 2
			i64.shr_u
			i64.const 3
			i64.const 2
			i64.rotl
			i64.const 1
			i64.const 65
			i64.shl
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
	// shl(1,4)=16, shr_s(16,2)=4, shr_u(16,2)=4
	// rotl(3,2)=12, shl(1,65)=shl(1,1)=2 (mod 64)
	assert.Len(t, results, 5)
	assert.Equal(t, int64(16), results[0])
	assert.Equal(t, int64(4), results[1])
	assert.Equal(t, int64(4), results[2])
	assert.Equal(t, int64(12), results[3])
	assert.Equal(t, int64(2), results[4])
}

func TestI32Extend8S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32)
			i32.const 127
			i32.extend8_s
			i32.const 128
			i32.extend8_s
			i32.const 255
			i32.extend8_s
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
	// 127 stays 127, 128 becomes -128, 255 becomes -1
	assert.Len(t, results, 3)
	assert.Equal(t, int32(127), results[0])
	assert.Equal(t, int32(-128), results[1])
	assert.Equal(t, int32(-1), results[2])
}

func TestI32Extend16S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32)
			i32.const 32767
			i32.extend16_s
			i32.const 32768
			i32.extend16_s
			i32.const 65535
			i32.extend16_s
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
	// 32767 stays 32767, 32768 becomes -32768, 65535 becomes -1
	assert.Len(t, results, 3)
	assert.Equal(t, int32(32767), results[0])
	assert.Equal(t, int32(-32768), results[1])
	assert.Equal(t, int32(-1), results[2])
}

func TestI32ReinterpretF32(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32)
			f32.const 0.0
			i32.reinterpret_f32
			f32.const 1.0
			i32.reinterpret_f32
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
	// 0.0 -> 0, 1.0 -> 1065353216
	assert.Len(t, results, 2)
	assert.Equal(t, int32(0), results[0])
	assert.Equal(t, int32(1065353216), results[1])
}

func TestI32TruncSatF32S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32)
			f32.const 3.7
			i32.trunc_sat_f32_s
			f32.const 0.0
			i32.trunc_sat_f32_s
			f32.const 1000000000000.0
			i32.trunc_sat_f32_s
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
	// 3.7 -> 3, 0.0 -> 0, overflow saturates to max
	assert.Len(t, results, 3)
	assert.Equal(t, int32(3), results[0])
	assert.Equal(t, int32(0), results[1])
	assert.Equal(t, int32(2147483647), results[2]) // MaxInt32
}

func TestI32WrapI64(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i32 i32 i32)
			i64.const 0
			i32.wrap_i64
			i64.const 100
			i32.wrap_i64
			i64.const 4294967296
			i32.wrap_i64
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
	// 0 -> 0, 100 -> 100, 4294967296 wraps to 0
	assert.Len(t, results, 3)
	assert.Equal(t, int32(0), results[0])
	assert.Equal(t, int32(100), results[1])
	assert.Equal(t, int32(0), results[2])
}

func TestI64Extend32S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64)
			i64.const 2147483647
			i64.extend32_s
			i64.const 2147483648
			i64.extend32_s
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
	// 2147483647 stays positive, 2147483648 becomes negative
	assert.Len(t, results, 2)
	assert.Equal(t, int64(2147483647), results[0])
	assert.Equal(t, int64(-2147483648), results[1])
}

func TestI64ExtendI32S(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64 i64)
			i32.const 100
			i64.extend_i32_s
			i32.const 0
			i64.extend_i32_s
			i32.const 2147483647
			i64.extend_i32_s
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
	// Sign extends i32 to i64
	assert.Len(t, results, 3)
	assert.Equal(t, int64(100), results[0])
	assert.Equal(t, int64(0), results[1])
	assert.Equal(t, int64(2147483647), results[2])
}

func TestI64ExtendI32U(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64 i64)
			i32.const 100
			i64.extend_i32_u
			i32.const 0
			i64.extend_i32_u
			i32.const 2147483647
			i64.extend_i32_u
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
	// Zero extends i32 to i64
	assert.Len(t, results, 3)
	assert.Equal(t, int64(100), results[0])
	assert.Equal(t, int64(0), results[1])
	assert.Equal(t, int64(2147483647), results[2])
}

func TestI64ReinterpretF64(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := testEnv.BuildWat(t, `
		(module
		  (func $subject (result i64 i64)
			f64.const 0.0
			i64.reinterpret_f64
			f64.const 1.0
			i64.reinterpret_f64
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
	// 0.0 -> 0, 1.0 -> 4607182418800017408
	assert.Len(t, results, 2)
	assert.Equal(t, int64(0), results[0])
	assert.Equal(t, int64(4607182418800017408), results[1])
}
