package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
	"wasp/cmd/goom/wasi_snapshot_preview1"
	"wasp/wasp"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		println("Usage: run <wasm file>")
		os.Exit(1)
	}

	wasm, err := os.ReadFile(args[0])
	if err != nil {
		println("Error reading file:", err.Error())
		os.Exit(1)
	}

	memstats(func() {
		start := time.Now()
		module, err := wasp.NewModule(wasm)
		if err != nil {
			println("Error loading module:", err.Error())
			os.Exit(1)
		}
		fmt.Println("WASM loaded in ", time.Since(start))

		linker := wasp.NewLinker()
		sp := wasi_snapshot_preview1.NewWasiSnapshotPreview1()
		sp.SetArgs(os.Args[1:]) // Pass remaining args to WASI
		sp.AddPreopen(3, ".")   // Preopen current directory as fd 3
		var linkerErrors = []error{
			linker.Define("wasi_snapshot_preview1", "args_get", sp.ArgsGet),
			linker.Define("wasi_snapshot_preview1", "args_sizes_get", sp.ArgsSizeGet),
			linker.Define("wasi_snapshot_preview1", "clock_res_get", sp.ClockResGet),
			linker.Define("wasi_snapshot_preview1", "clock_time_get", sp.ClockTimeGet),
			linker.Define("wasi_snapshot_preview1", "fd_close", sp.FdClose),
			linker.Define("wasi_snapshot_preview1", "fd_fdstat_get", sp.FdStatGet),
			linker.Define("wasi_snapshot_preview1", "fd_fdstat_set_flags", sp.FdStatSetFlags),
			linker.Define("wasi_snapshot_preview1", "fd_filestat_get", sp.FdFilestatGet),
			linker.Define("wasi_snapshot_preview1", "fd_filestat_set_size", sp.FdFilestatSetSize),
			linker.Define("wasi_snapshot_preview1", "fd_filestat_set_times", sp.FdFilestatSetTimes),
			linker.Define("wasi_snapshot_preview1", "fd_pread", sp.FdPRead),
			linker.Define("wasi_snapshot_preview1", "fd_prestat_get", sp.FdPrestatGet),
			linker.Define("wasi_snapshot_preview1", "fd_prestat_dir_name", sp.FdPrestatDirName),
			linker.Define("wasi_snapshot_preview1", "fd_pwrite", sp.FdPWrite),
			linker.Define("wasi_snapshot_preview1", "fd_read", sp.FdRead),
			linker.Define("wasi_snapshot_preview1", "fd_seek", sp.FdSeek),
			linker.Define("wasi_snapshot_preview1", "fd_write", sp.FdWrite),
			linker.Define("wasi_snapshot_preview1", "path_create_directory", sp.PathCreateDirectory),
			linker.Define("wasi_snapshot_preview1", "path_filestat_get", sp.PathFilestatGet),
			linker.Define("wasi_snapshot_preview1", "path_filestat_set_times", sp.PathFilestatSetTimes),
			linker.Define("wasi_snapshot_preview1", "path_link", sp.PathLink),
			linker.Define("wasi_snapshot_preview1", "path_open", sp.PathOpen),
			linker.Define("wasi_snapshot_preview1", "path_readlink", sp.PathReadLink),
			linker.Define("wasi_snapshot_preview1", "path_remove_directory", sp.PathRemoveDirectory),
			linker.Define("wasi_snapshot_preview1", "path_rename", sp.PathRename),
			linker.Define("wasi_snapshot_preview1", "path_symlink", sp.PathSymlink),
			linker.Define("wasi_snapshot_preview1", "path_unlink_file", sp.PathUnlinkFile),
			linker.Define("wasi_snapshot_preview1", "proc_exit", sp.ProcExit),
			linker.Define("wasi_snapshot_preview1", "random_get", sp.RandomGet),
			linker.Define("wasi_snapshot_preview1", "fd_readdir", sp.FdReadDir),
			linker.Define("wasi_snapshot_preview1", "fd_sync", sp.FdSync),
			linker.Define("wasi_snapshot_preview1", "poll_oneoff", sp.PollOneOf),
		}
		for _, err := range linkerErrors {
			if err != nil {
				println("Error defining function:", err.Error())
				os.Exit(1)
			}
		}

		instance, err := wasp.NewInstance(
			module,
			wasp.WithLinker(linker),
		)
		if err != nil {
			println("Error creating runtime:", err.Error())
			os.Exit(1)
		}

		_ = instance

		fmt.Println("imports:")
		for _, imp := range module.Imports() {
			fmt.Printf("\t %s\n", imp.String())
		}

		fmt.Println("exports:")
		for name, exp := range module.Exports() {
			fmt.Printf("\t %s (kind: %d)\n", name, exp.Kind())
		}

		fn, err := module.GetExportedFunction("_start")
		if err != nil {
			println("Error getting function:", err.Error())
			os.Exit(1)
		}

		var results []any
		elapsed(func() {
			_, results, err = instance.Call(fn)
			if err != nil {
				println("Error calling function:", err.Error())
				os.Exit(1)
			}
		})

		fmt.Printf("Results: %v\n", results)
	})
}

func memstats(f func()) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)

	fmt.Printf("Memory usage: %d bytes\n", after.Alloc-before.Alloc)
	fmt.Printf("Heap allocations: %d bytes\n", after.HeapAlloc-before.HeapAlloc)
	fmt.Printf("Heap objects: %d\n", after.HeapObjects-before.HeapObjects)
}

func elapsed(f func()) {
	start := time.Now()
	f()
	fmt.Printf("Elapsed time: %s\n", time.Since(start))
}
