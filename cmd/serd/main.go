package main

import (
	"errors"
	"fmt"
	"os"
	"time"
	"wasp/pkg/images"
	"wasp/wasi"
	"wasp/wasp"
)

const (
	dumpFileName = "dump.bin"
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

	var store *wasp.Store
	var instance *wasp.Instance

	linker := wasp.NewLinker()
	sp := wasi.NewWasiSnapshotPreview1()
	sp.SetArgs(nil)       // Pass remaining args to WASI
	sp.AddPreopen(3, ".") // Preopen current directory as fd 3
	if err := sp.Register(linker); err != nil {
		panic(err)
	}
	linker.Define("env", "save_film", func(ptr, height, width int32) {
		size := height * width * 3
		data := store.Memories[0].Load(int(ptr), int(size))

		err := images.SaveBMP("output.bmp", data, int(width), int(height))
		if err != nil {
			println("Error saving BMP:", err.Error())
			os.Exit(1)
		}
	})

	var hasDump bool
	dump, err := os.Open(dumpFileName)
	if err == nil {
		start := time.Now()
		store, instance, err = wasp.DeserializeState(dump, module, linker)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Deserialization took: %v\n", time.Since(start))
		hasDump = true
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			panic(err)
		}
	}
	defer dump.Close()

	if !hasDump {
		store = wasp.NewStore(module)
		instance, err = wasp.NewInstance(module, store, wasp.WithLinker(linker))
		if err != nil {
			panic(err)
		}
		funcref, err := module.GetExportedFunction("_start")
		if err != nil {
			panic(err)
		}

		_, err = instance.Call(funcref)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("hasDump: %v\n", hasDump)

	sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use

	go func() {
		if hasDump {
			return
		}
		time.Sleep(15 * time.Second)
		instance.Pause()
	}()

	if err := instance.Run(); err != nil {
		panic(err)
	}

	if !hasDump {
		file, err := os.Create(dumpFileName)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		start := time.Now()
		err = wasp.SerializeState(file, store, instance)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Serialization took: %v\n", time.Since(start))
	}
}
