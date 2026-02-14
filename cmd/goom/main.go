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
		var linkerErrors = []error{
			linker.Define("wasi_snapshot_preview1", "args_get", wasi_snapshot_preview1.ArgsGet),
			linker.Define("wasi_snapshot_preview1", "args_sizes_get", wasi_snapshot_preview1.ArgsSizeGet),
			linker.Define("wasi_snapshot_preview1", "clock_res_get", wasi_snapshot_preview1.ClockResGet),
			linker.Define("wasi_snapshot_preview1", "clock_time_get", wasi_snapshot_preview1.ClockTimeGet),
			linker.Define("wasi_snapshot_preview1", "fd_close", wasi_snapshot_preview1.FdClose),
			linker.Define("wasi_snapshot_preview1", "fd_fdstat_get", wasi_snapshot_preview1.FdStatGet),
			linker.Define("wasi_snapshot_preview1", "fd_fdstat_set_flags", wasi_snapshot_preview1.FdStatSetFlags),
			linker.Define("wasi_snapshot_preview1", "fd_filestat_get", wasi_snapshot_preview1.FdFilestatGet),
			linker.Define("wasi_snapshot_preview1", "fd_filestat_set_size", wasi_snapshot_preview1.FdFilestatSetSize),
			linker.Define("wasi_snapshot_preview1", "fd_filestat_set_times", wasi_snapshot_preview1.FdFilestatSetTimes),
			linker.Define("wasi_snapshot_preview1", "fd_pread", wasi_snapshot_preview1.FdPRead),
			linker.Define("wasi_snapshot_preview1", "fd_prestat_get", wasi_snapshot_preview1.FdPrestatGet),
			linker.Define("wasi_snapshot_preview1", "fd_prestat_dir_name", wasi_snapshot_preview1.FdPrestatDirName),
			linker.Define("wasi_snapshot_preview1", "fd_pwrite", wasi_snapshot_preview1.FdPWrite),
			linker.Define("wasi_snapshot_preview1", "fd_read", wasi_snapshot_preview1.FdRead),
			linker.Define("wasi_snapshot_preview1", "fd_seek", wasi_snapshot_preview1.FdSeek),
			linker.Define("wasi_snapshot_preview1", "fd_write", wasi_snapshot_preview1.FdWrite),
			linker.Define("wasi_snapshot_preview1", "path_create_directory", wasi_snapshot_preview1.PathCreateDirectory),
			linker.Define("wasi_snapshot_preview1", "path_filestat_get", wasi_snapshot_preview1.PathFilestatGet),
			linker.Define("wasi_snapshot_preview1", "path_filestat_set_times", wasi_snapshot_preview1.PathFilestatSetTimes),
			linker.Define("wasi_snapshot_preview1", "path_link", wasi_snapshot_preview1.PathLink),
			linker.Define("wasi_snapshot_preview1", "path_open", wasi_snapshot_preview1.PathOpen),
			linker.Define("wasi_snapshot_preview1", "path_readlink", wasi_snapshot_preview1.PathReadLink),
			linker.Define("wasi_snapshot_preview1", "path_remove_directory", wasi_snapshot_preview1.PathRemoveDirectory),
			linker.Define("wasi_snapshot_preview1", "path_rename", wasi_snapshot_preview1.PathRename),
			linker.Define("wasi_snapshot_preview1", "path_symlink", wasi_snapshot_preview1.PathSymlink),
			linker.Define("wasi_snapshot_preview1", "path_unlink_file", wasi_snapshot_preview1.PathUnlinkFile),
			linker.Define("wasi_snapshot_preview1", "proc_exit", wasi_snapshot_preview1.ProcExit),
			linker.Define("wasi_snapshot_preview1", "random_get", wasi_snapshot_preview1.RandomGet),
			linker.Define("wasi_snapshot_preview1", "fd_readdir", wasi_snapshot_preview1.FdReadDir),
			linker.Define("wasi_snapshot_preview1", "fd_sync", wasi_snapshot_preview1.FdSync),
			linker.Define("wasi_snapshot_preview1", "poll_oneoff", wasi_snapshot_preview1.PollOneOf),
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
