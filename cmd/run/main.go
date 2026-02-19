package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"wasp/wasi"
	"wasp/wasp"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		println("Usage: run <wasm file>")
		os.Exit(1)
	}

	wasm, err := os.ReadFile(args[0])
	if err != nil {
		println("Error reading file:", err.Error())
		os.Exit(1)
	}

	module, err := wasp.NewModule(wasm)
	if err != nil {
		println("Error loading module:", err.Error())
		os.Exit(1)
	}

	var requiresWasi bool
	for _, imp := range module.Imports() {
		if imp.ModuleName == "wasi_snapshot_preview1" {
			requiresWasi = true
			break
		}
	}

	store := wasp.NewStore(module)
	defaultMem := store.Memories[0]
	defaultMem.Grow(200)

	linker := wasp.NewLinker()

	linker.Define("env", "save_film", func(ptr, height, width int32) {
		size := height * width * 3
		data := defaultMem.Load(int(ptr), int(size))

		saveBin(data)

		err := saveBMP("output.bmp", data, int(width), int(height))
		if err != nil {
			println("Error saving BMP:", err.Error())
			os.Exit(1)
		}
	})

	if requiresWasi {
		sp := wasi.NewWasiSnapshotPreview1()
		sp.SetArgs(args)                // Pass remaining args to WASI
		sp.AddPreopen(3, ".")           // Preopen current directory as fd 3
		sp.SetMemory(store.Memories[0]) // Set the memory for WASI to use
		if err := sp.Register(linker); err != nil {
			println("Error registering WASI snapshot preview 1:", err.Error())
			os.Exit(1)
		}
	}

	var autoRun = true
	var runFunc string
	options := []wasp.InstanceOption{
		wasp.WithLinker(linker),
	}
	for i, arg := range args {
		switch arg {
		case "--verbose", "-v":
			options = append(options, wasp.Verbose())
		case "--dry-run":
			autoRun = false
		case "--func", "-f":
			runFunc = args[i+1]
		}
	}

	instance, err := wasp.NewInstance(
		module,
		store,
		options...,
	)
	if err != nil {
		println("Error creating instance of module:", err.Error())
		os.Exit(1)
	}

	if autoRun {
		funcref, err := findCandidateForStartFunc(module, runFunc)
		if err != nil {
			println("Error getting start function:", err.Error())
			os.Exit(1)
		}

		if _, err = instance.Call(funcref); err != nil {
			println("Error calling start function:", err.Error())
			os.Exit(1)
		}

		if err := instance.Tick(); err != nil {
			println("Error ticking instance:", err.Error())
			os.Exit(1)
		}
	}
}

func saveBin(data []byte) {
	if err := os.WriteFile("output.bin", data, os.ModePerm); err != nil {
		println("Error saving binary data:", err.Error())
		os.Exit(1)
	}
}

func findCandidateForStartFunc(module *wasp.Module, target string) (int, error) {
	if target != "" {
		return module.GetExportedFunction(target)
	}

	if fn, err := module.GetStartFunction(); err == nil {
		return fn, nil
	}

	if fn, err := module.GetExportedFunction("_start"); err == nil {
		return fn, nil
	}

	return -1, fmt.Errorf("could not find candidate for start function")
}

func saveBMP(filename string, data []byte, width, height int) error {
	// BMP files require rows to be padded to 4-byte boundaries
	rowSize := (width*3 + 3) &^ 3
	pixelDataSize := rowSize * height

	// BMP file header (14 bytes) + DIB header (40 bytes) = 54 bytes
	fileSize := 54 + pixelDataSize

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// BMP File Header (14 bytes)
	file.Write([]byte{'B', 'M'})                              // Signature
	binary.Write(file, binary.LittleEndian, uint32(fileSize)) // File size
	binary.Write(file, binary.LittleEndian, uint16(0))        // Reserved
	binary.Write(file, binary.LittleEndian, uint16(0))        // Reserved
	binary.Write(file, binary.LittleEndian, uint32(54))       // Pixel data offset

	// DIB Header (BITMAPINFOHEADER - 40 bytes)
	binary.Write(file, binary.LittleEndian, uint32(40))            // DIB header size
	binary.Write(file, binary.LittleEndian, int32(width))          // Width
	binary.Write(file, binary.LittleEndian, int32(height))         // Height (positive = bottom-up)
	binary.Write(file, binary.LittleEndian, uint16(1))             // Color planes
	binary.Write(file, binary.LittleEndian, uint16(24))            // Bits per pixel
	binary.Write(file, binary.LittleEndian, uint32(0))             // Compression (none)
	binary.Write(file, binary.LittleEndian, uint32(pixelDataSize)) // Image size
	binary.Write(file, binary.LittleEndian, int32(2835))           // Horizontal resolution (72 DPI)
	binary.Write(file, binary.LittleEndian, int32(2835))           // Vertical resolution (72 DPI)
	binary.Write(file, binary.LittleEndian, uint32(0))             // Colors in palette
	binary.Write(file, binary.LittleEndian, uint32(0))             // Important colors

	// Pixel data (BMP stores rows bottom-to-top and BGR format)
	padding := make([]byte, rowSize-width*3)
	for y := height - 1; y >= 0; y-- {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 3
			// Convert RGB to BGR
			file.Write([]byte{data[i+2], data[i+1], data[i]})
		}
		file.Write(padding)
	}

	return nil
}
