package main_test

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
	"wasp/wasi"
	"wasp/wasp"
)

var instance *wasp.Instance
var fn int

func init() {
	wasm, err := os.ReadFile("dg.wasm")
	if err != nil {
		panic(err)
	}

	module, err := wasp.NewModule(wasm)
	if err != nil {
		panic(err)
	}

	store := wasp.NewStore(module)

	linker := wasp.NewLinker()

	dg(linker)

	sp := wasi.NewWasiSnapshotPreview1()
	sp.SetArgs([]string{"doom1.wad"}) // Pass remaining args to WASI
	sp.AddPreopen(3, ".")             // Preopen current directory as fd 3
	//sp.AddPreopen(4, "doom1.wad")
	sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use
	if err := sp.Register(linker); err != nil {
		panic(err)
	}

	options := []wasp.InstanceOption{
		wasp.WithLinker(linker),
		wasp.IgnoreUnreachable(), // Allow DOOM to continue despite UBSan panics
	}

	instance, err = wasp.NewInstance(
		module,
		store,
		options...,
	)
	if err != nil {
		println("Error creating runtime:", err.Error())
		os.Exit(1)
	}

	fn, err = module.GetExportedFunction("_start")
	if err != nil {
		println("Error getting function:", err.Error())
		os.Exit(1)
	}
}

func TestGoom(t *testing.T) {
	_, err := instance.Call(fn)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(10 * time.Second)
		instance.Pause()
	}()

	if err := instance.Run(); err != nil {
		t.Fatal(err)
	}
}

func dg(linker *wasp.Linker) {
	var started time.Time
	var init sync.Once

	linker.Define("dg", "init", func(resx, resy int32) {
		init.Do(func() {
			started = time.Now()
		})
	})

	linker.Define("dg", "draw_frame", func(ptr int32) {
		return
	})

	linker.Define("dg", "sleep_ms", func(ms int32) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	})

	linker.Define("dg", "get_ticks_ms", func() int32 {
		elapsed := time.Since(started)
		ms := int32(elapsed.Milliseconds())
		return ms
	})

	linker.Define("dg", "get_key", func() int32 {
		return -1
	})

	linker.Define("dg", "set_window_title", func(ptr, len int32) {
		fmt.Printf("DB set_window_title called with ptr=%d, len=%d\n", ptr, len)
	})

	// UBSan stubs - allow undefined behavior to continue without panicking
	linker.Define("env", "__ubsan_handle_shift_out_of_bounds", func(dataPtr, lhs, rhs int32) {
		// Log but don't panic - this allows DOOM to continue despite UB
		fmt.Printf("[UBSAN] shift out of bounds: lhs=%d, rhs=%d (continuing)\n", lhs, rhs)
	})
}
