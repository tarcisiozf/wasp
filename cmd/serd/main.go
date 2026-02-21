package main

import (
	"fmt"
	"os"
	"time"
	"wasp/wasi"
	"wasp/wasp"
)

func main() {
	wasm, err := os.ReadFile("ray.wasm")
	if err != nil {
		panic(err)
	}

	module, err := wasp.NewModule(wasm)
	if err != nil {
		panic(err)
	}

	store := wasp.NewStore(module)

	linker := wasp.NewLinker()
	sp := wasi.NewWasiSnapshotPreview1()
	sp.SetArgs(nil)                 // Pass remaining args to WASI
	sp.AddPreopen(3, ".")           // Preopen current directory as fd 3
	sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use
	if err := sp.Register(linker); err != nil {
		panic(err)
	}
	linker.Define("env", "save_film", func(ptr, h, w int32) {})

	funcref, err := module.GetExportedFunction("_start")
	if err != nil {
		panic(err)
	}

	instance, err := wasp.NewInstance(module, store, wasp.WithLinker(linker))
	if err != nil {
		panic(err)
	}

	_, err = instance.Call(funcref)
	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(3 * time.Second)
		instance.Pause()
	}()

	if err := instance.Run(); err != nil {
		panic(err)
	}

	file, err := os.Create("dump.bin")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = wasp.Foo(file, store, instance)
	if err != nil {
		panic(err)
	}

	fmt.Println("@ END")
}
