package instructions_test

import (
	"testing"
	"wasp/wasp/tests"

	"github.com/stretchr/testify/assert"
)

func TestConst(t *testing.T) {
	testEnv := tests.NewEnvironment()
	build, err := tests.BuildWat(t, `
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
