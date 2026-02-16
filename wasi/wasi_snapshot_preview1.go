package wasi

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
	"wasp/wasp"
)

// WASI error codes
const (
	ErrnoSuccess    int32 = 0  // No error occurred
	ErrnoBadf       int32 = 8  // Bad file descriptor
	ErrnoInval      int32 = 28 // Invalid argument
	ErrnoNosys      int32 = 52 // Function not supported
	ErrnoNoent      int32 = 44 // No such file or directory
	ErrnoNotdir     int32 = 54 // Not a directory
	ErrnoIsdir      int32 = 31 // Is a directory
	ErrnoExist      int32 = 20 // File exists
	ErrnoAccess     int32 = 2  // Permission denied
	ErrnoAgain      int32 = 6  // Resource unavailable
	ErrnoNotcapable int32 = 76 // Capabilities insufficient
	ErrnoIo         int32 = 29 // I/O error
	ErrnoFault      int32 = 21 // Bad address
)

// File descriptor types
const (
	FdStdin  int32 = 0
	FdStdout int32 = 1
	FdStderr int32 = 2
	FdFirst  int32 = 3 // First available fd for opened files
)

// Filetype constants
const (
	FiletypeUnknown         uint8 = 0
	FiletypeBlockDevice     uint8 = 1
	FiletypeCharacterDevice uint8 = 2
	FiletypeDirectory       uint8 = 3
	FiletypeRegularFile     uint8 = 4
	FiletypeSocketDgram     uint8 = 5
	FiletypeSocketStream    uint8 = 6
	FiletypeSymbolicLink    uint8 = 7
)

// Clock IDs
const (
	ClockRealtime         int32 = 0
	ClockMonotonic        int32 = 1
	ClockProcessCputimeId int32 = 2
	ClockThreadCputimeId  int32 = 3
)

// Whence constants for seek
const (
	WhenceSet int32 = 0
	WhenceCur int32 = 1
	WhenceEnd int32 = 2
)

// Memory interface for WASI to read/write WebAssembly linear memory
type Memory interface {
	Load(offset int, size int) []byte
	Store(offset int, data []byte)
}

// FileDescriptor represents an open file
type FileDescriptor struct {
	File     *os.File
	Path     string
	Filetype uint8
	Offset   int64
}

// WasiSnapshotPreview1 implements WASI preview1 interface
type WasiSnapshotPreview1 struct {
	memory    Memory
	args      []string
	env       []string
	preopens  map[int32]string // fd -> path
	openFiles map[int32]*FileDescriptor
	nextFd    int32
	exitCode  int32
	hasExited bool
}

// NewWasiSnapshotPreview1 creates a new WASI instance
func NewWasiSnapshotPreview1() *WasiSnapshotPreview1 {
	return &WasiSnapshotPreview1{
		args:      os.Args,
		env:       os.Environ(),
		preopens:  make(map[int32]string),
		openFiles: make(map[int32]*FileDescriptor),
		nextFd:    FdFirst,
	}
}

// SetMemory sets the memory interface
func (sp *WasiSnapshotPreview1) SetMemory(mem Memory) {
	sp.memory = mem
}

// SetArgs sets command line arguments
func (sp *WasiSnapshotPreview1) SetArgs(args []string) {
	sp.args = args
}

// SetEnv sets environment variables
func (sp *WasiSnapshotPreview1) SetEnv(env []string) {
	sp.env = env
}

// AddPreopen adds a preopened directory
func (sp *WasiSnapshotPreview1) AddPreopen(fd int32, path string) {
	sp.preopens[fd] = path
	if fd >= sp.nextFd {
		sp.nextFd = fd + 1
	}
}

// writeUint32 writes a uint32 to memory at the given offset
func (sp *WasiSnapshotPreview1) writeUint32(offset int32, value uint32) {
	if sp.memory == nil {
		return
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, value)
	sp.memory.Store(int(offset), buf)
}

// writeUint64 writes a uint64 to memory at the given offset
func (sp *WasiSnapshotPreview1) writeUint64(offset int32, value uint64) {
	if sp.memory == nil {
		return
	}
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, value)
	sp.memory.Store(int(offset), buf)
}

// readUint32 reads a uint32 from memory at the given offset
func (sp *WasiSnapshotPreview1) readUint32(offset int32) uint32 {
	if sp.memory == nil {
		return 0
	}
	buf := sp.memory.Load(int(offset), 4)
	return binary.LittleEndian.Uint32(buf)
}

// readUint64 reads a uint64 from memory at the given offset
func (sp *WasiSnapshotPreview1) readUint64(offset int32) uint64 {
	if sp.memory == nil {
		return 0
	}
	buf := sp.memory.Load(int(offset), 8)
	return binary.LittleEndian.Uint64(buf)
}

// readBytes reads bytes from memory
func (sp *WasiSnapshotPreview1) readBytes(offset int32, length int32) []byte {
	if sp.memory == nil {
		return nil
	}
	return sp.memory.Load(int(offset), int(length))
}

// writeBytes writes bytes to memory
func (sp *WasiSnapshotPreview1) writeBytes(offset int32, data []byte) {
	if sp.memory == nil {
		return
	}
	sp.memory.Store(int(offset), data)
}

// ArgsGet writes the argument pointers and argument strings to memory
func (sp *WasiSnapshotPreview1) ArgsGet(argv int32, argvBuf int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	bufOffset := argvBuf
	for i, arg := range sp.args {
		// Write pointer to this argument
		sp.writeUint32(argv+int32(i*4), uint32(bufOffset))

		// Write the argument string with null terminator
		argBytes := append([]byte(arg), 0)
		sp.writeBytes(bufOffset, argBytes)
		bufOffset += int32(len(argBytes))
	}

	return ErrnoSuccess
}

// ArgsSizeGet returns the number of arguments and the total buffer size needed
func (sp *WasiSnapshotPreview1) ArgsSizeGet(argc int32, argvBufSize int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	sp.writeUint32(argc, uint32(len(sp.args)))

	totalSize := 0
	for _, arg := range sp.args {
		totalSize += len(arg) + 1 // +1 for null terminator
	}
	sp.writeUint32(argvBufSize, uint32(totalSize))

	return ErrnoSuccess
}

// ClockResGet returns the resolution of the specified clock
func (sp *WasiSnapshotPreview1) ClockResGet(clockId int32, resolution int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	var res uint64
	switch clockId {
	case ClockRealtime, ClockMonotonic:
		res = 1 // 1 nanosecond resolution
	case ClockProcessCputimeId, ClockThreadCputimeId:
		res = 1000 // 1 microsecond resolution
	default:
		return ErrnoInval
	}

	sp.writeUint64(resolution, res)
	return ErrnoSuccess
}

// ClockTimeGet returns the current time for the specified clock
func (sp *WasiSnapshotPreview1) ClockTimeGet(clockId int32, precision int64, timestamp int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	var t uint64
	switch clockId {
	case ClockRealtime:
		t = uint64(time.Now().UnixNano())
	case ClockMonotonic:
		t = uint64(time.Now().UnixNano())
	case ClockProcessCputimeId, ClockThreadCputimeId:
		t = uint64(time.Now().UnixNano())
	default:
		return ErrnoInval
	}

	sp.writeUint64(timestamp, t)
	return ErrnoSuccess
}

// FdClose closes a file descriptor
func (sp *WasiSnapshotPreview1) FdClose(fd int32) int32 {
	// Standard streams can't be closed
	if fd < FdFirst {
		return ErrnoBadf
	}

	if fileDesc, ok := sp.openFiles[fd]; ok {
		if fileDesc.File != nil {
			fileDesc.File.Close()
		}
		delete(sp.openFiles, fd)
		return ErrnoSuccess
	}

	// Check if it's a preopen
	if _, ok := sp.preopens[fd]; ok {
		delete(sp.preopens, fd)
		return ErrnoSuccess
	}

	return ErrnoBadf
}

// FdStatGet returns the fdstat for a file descriptor
func (sp *WasiSnapshotPreview1) FdStatGet(fd int32, statPtr int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	var filetype uint8
	var rights uint64 = 0xFFFFFFFFFFFFFFFF // All rights
	var rightsInheriting uint64 = 0xFFFFFFFFFFFFFFFF

	switch fd {
	case FdStdin:
		filetype = FiletypeCharacterDevice
	case FdStdout, FdStderr:
		filetype = FiletypeCharacterDevice
	default:
		if _, ok := sp.preopens[fd]; ok {
			filetype = FiletypeDirectory
		} else if fileDesc, ok := sp.openFiles[fd]; ok {
			filetype = fileDesc.Filetype
		} else {
			return ErrnoBadf
		}
	}

	// fdstat structure:
	// - fs_filetype: u8 (offset 0)
	// - fs_flags: u16 (offset 2)
	// - fs_rights_base: u64 (offset 8)
	// - fs_rights_inheriting: u64 (offset 16)
	sp.writeBytes(statPtr, []byte{filetype, 0}) // filetype + padding
	sp.writeUint32(statPtr+2, 0)                // flags + padding
	sp.writeUint64(statPtr+8, rights)
	sp.writeUint64(statPtr+16, rightsInheriting)

	return ErrnoSuccess
}

// FdStatSetFlags sets the flags for a file descriptor
func (sp *WasiSnapshotPreview1) FdStatSetFlags(fd int32, flags int32) int32 {
	// Validate fd
	if fd < 0 {
		return ErrnoBadf
	}
	// Setting flags is not commonly implemented, return success
	return ErrnoSuccess
}

// FdFilestatGet returns file information for a file descriptor
func (sp *WasiSnapshotPreview1) FdFilestatGet(fd int32, filestat int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	var stat os.FileInfo
	var filetype uint8

	switch fd {
	case FdStdin:
		stat, _ = os.Stdin.Stat()
		filetype = FiletypeCharacterDevice
	case FdStdout:
		stat, _ = os.Stdout.Stat()
		filetype = FiletypeCharacterDevice
	case FdStderr:
		stat, _ = os.Stderr.Stat()
		filetype = FiletypeCharacterDevice
	default:
		if path, ok := sp.preopens[fd]; ok {
			var err error
			stat, err = os.Stat(path)
			if err != nil {
				return ErrnoIo
			}
			filetype = FiletypeDirectory
		} else if fileDesc, ok := sp.openFiles[fd]; ok {
			var err error
			stat, err = fileDesc.File.Stat()
			if err != nil {
				return ErrnoIo
			}
			filetype = fileDesc.Filetype
		} else {
			return ErrnoBadf
		}
	}

	// Write filestat structure
	sp.writeUint64(filestat, 0)                  // dev
	sp.writeUint64(filestat+8, 0)                // ino
	sp.writeBytes(filestat+16, []byte{filetype}) // filetype
	sp.writeUint64(filestat+24, 1)               // nlink
	if stat != nil {
		sp.writeUint64(filestat+32, uint64(stat.Size()))               // size
		sp.writeUint64(filestat+40, uint64(stat.ModTime().UnixNano())) // atim
		sp.writeUint64(filestat+48, uint64(stat.ModTime().UnixNano())) // mtim
		sp.writeUint64(filestat+56, uint64(stat.ModTime().UnixNano())) // ctim
	}

	return ErrnoSuccess
}

// FdFilestatSetSize changes the size of an open file
func (sp *WasiSnapshotPreview1) FdFilestatSetSize(fd int32, size int64) int32 {
	if fileDesc, ok := sp.openFiles[fd]; ok && fileDesc.File != nil {
		if err := fileDesc.File.Truncate(size); err != nil {
			return ErrnoIo
		}
		return ErrnoSuccess
	}

	return ErrnoBadf
}

// FdFilestatSetTimes sets the access and modification times
func (sp *WasiSnapshotPreview1) FdFilestatSetTimes(fd int32, atim int64, mtim int64, fstFlags int32) int32 {
	var path string
	if p, ok := sp.preopens[fd]; ok {
		path = p
	} else if fileDesc, ok := sp.openFiles[fd]; ok {
		path = fileDesc.Path
	} else {
		return ErrnoBadf
	}

	atime := time.Unix(0, atim)
	mtime := time.Unix(0, mtim)

	if err := os.Chtimes(path, atime, mtime); err != nil {
		return ErrnoIo
	}

	return ErrnoSuccess
}

// FdPRead reads from a file at a specific offset without changing the fd position
func (sp *WasiSnapshotPreview1) FdPRead(fd int32, iovs int32, iovsLen int32, offset int64, nread int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	fileDesc, ok := sp.openFiles[fd]

	if !ok || fileDesc.File == nil {
		return ErrnoBadf
	}

	totalRead := uint32(0)
	for i := int32(0); i < iovsLen; i++ {
		iovPtr := iovs + i*8
		bufPtr := int32(sp.readUint32(iovPtr))
		bufLen := int32(sp.readUint32(iovPtr + 4))

		buf := make([]byte, bufLen)
		n, err := fileDesc.File.ReadAt(buf, offset+int64(totalRead))
		if n > 0 {
			sp.writeBytes(bufPtr, buf[:n])
			totalRead += uint32(n)
		}
		if err != nil {
			break
		}
	}

	sp.writeUint32(nread, totalRead)
	return ErrnoSuccess
}

// FdPrestatGet returns the prestat for a file descriptor
func (sp *WasiSnapshotPreview1) FdPrestatGet(fd int32, prestat int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	path, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	// prestat structure:
	// - tag: u8 (0 = directory)
	// - u.dir.pr_name_len: u32
	sp.writeBytes(prestat, []byte{0}) // tag = 0 (directory)
	sp.writeUint32(prestat+4, uint32(len(path)))

	return ErrnoSuccess
}

// FdPrestatDirName returns the path of a preopened directory
func (sp *WasiSnapshotPreview1) FdPrestatDirName(fd int32, path int32, pathLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	preopenPath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	if int32(len(preopenPath)) > pathLen {
		return ErrnoInval
	}

	sp.writeBytes(path, []byte(preopenPath))

	return ErrnoSuccess
}

// FdPWrite writes to a file at a specific offset without changing the fd position
func (sp *WasiSnapshotPreview1) FdPWrite(fd int32, iovs int32, iovsLen int32, offset int64, nwritten int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	fileDesc, ok := sp.openFiles[fd]

	if !ok || fileDesc.File == nil {
		return ErrnoBadf
	}

	totalWritten := uint32(0)
	for i := int32(0); i < iovsLen; i++ {
		iovPtr := iovs + i*8
		bufPtr := int32(sp.readUint32(iovPtr))
		bufLen := int32(sp.readUint32(iovPtr + 4))

		buf := sp.readBytes(bufPtr, bufLen)
		n, err := fileDesc.File.WriteAt(buf, offset+int64(totalWritten))
		totalWritten += uint32(n)
		if err != nil {
			break
		}
	}

	sp.writeUint32(nwritten, totalWritten)
	return ErrnoSuccess
}

// FdRead reads from a file descriptor
func (sp *WasiSnapshotPreview1) FdRead(fd int32, iovs int32, iovsLen int32, nread int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	var reader io.Reader
	switch fd {
	case FdStdin:
		reader = os.Stdin
	default:
		fileDesc, ok := sp.openFiles[fd]
		if !ok || fileDesc.File == nil {
			return ErrnoBadf
		}
		reader = fileDesc.File
	}

	totalRead := uint32(0)
	for i := int32(0); i < iovsLen; i++ {
		iovPtr := iovs + i*8
		bufPtr := int32(sp.readUint32(iovPtr))
		bufLen := int32(sp.readUint32(iovPtr + 4))

		buf := make([]byte, bufLen)
		n, err := reader.Read(buf)
		if n > 0 {
			sp.writeBytes(bufPtr, buf[:n])
			totalRead += uint32(n)
		}
		if err != nil {
			break
		}
	}

	sp.writeUint32(nread, totalRead)
	return ErrnoSuccess
}

// FdSeek changes the offset of a file descriptor
func (sp *WasiSnapshotPreview1) FdSeek(fd int32, offset int64, whence int32, newoffset int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	fileDesc, ok := sp.openFiles[fd]

	if !ok || fileDesc.File == nil {
		return ErrnoBadf
	}

	var seekWhence int
	switch whence {
	case WhenceSet:
		seekWhence = io.SeekStart
	case WhenceCur:
		seekWhence = io.SeekCurrent
	case WhenceEnd:
		seekWhence = io.SeekEnd
	default:
		return ErrnoInval
	}

	newPos, err := fileDesc.File.Seek(offset, seekWhence)
	if err != nil {
		return ErrnoIo
	}

	sp.writeUint64(newoffset, uint64(newPos))
	return ErrnoSuccess
}

// FdWrite writes to a file descriptor
func (sp *WasiSnapshotPreview1) FdWrite(fd int32, iovs int32, iovsLen int32, nwritten int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	var writer io.Writer
	switch fd {
	case FdStdout:
		writer = os.Stdout
	case FdStderr:
		writer = os.Stderr
	default:
		fileDesc, ok := sp.openFiles[fd]
		if !ok || fileDesc.File == nil {
			return ErrnoBadf
		}
		writer = fileDesc.File
	}

	totalWritten := uint32(0)
	for i := int32(0); i < iovsLen; i++ {
		iovPtr := iovs + i*8
		bufPtr := int32(sp.readUint32(iovPtr))
		bufLen := int32(sp.readUint32(iovPtr + 4))

		buf := sp.readBytes(bufPtr, bufLen)
		n, err := writer.Write(buf)
		totalWritten += uint32(n)
		if err != nil {
			break
		}
	}

	sp.writeUint32(nwritten, totalWritten)
	return ErrnoSuccess
}

// PathCreateDirectory creates a directory
func (sp *WasiSnapshotPreview1) PathCreateDirectory(fd int32, pathPtr int32, pathLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	pathBytes := sp.readBytes(pathPtr, pathLen)
	fullPath := basePath + "/" + string(pathBytes)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		if os.IsExist(err) {
			return ErrnoExist
		}
		return ErrnoIo
	}

	return ErrnoSuccess
}

// PathFilestatGet returns file information for a path
func (sp *WasiSnapshotPreview1) PathFilestatGet(fd int32, flags int32, pathPtr int32, pathLen int32, filestat int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	pathBytes := sp.readBytes(pathPtr, pathLen)
	fullPath := basePath + "/" + string(pathBytes)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrnoNoent
		}
		return ErrnoIo
	}

	var filetype uint8 = FiletypeRegularFile
	if stat.IsDir() {
		filetype = FiletypeDirectory
	}

	// Write filestat structure
	sp.writeUint64(filestat, 0)                                    // dev
	sp.writeUint64(filestat+8, 0)                                  // ino
	sp.writeBytes(filestat+16, []byte{filetype})                   // filetype
	sp.writeUint64(filestat+24, 1)                                 // nlink
	sp.writeUint64(filestat+32, uint64(stat.Size()))               // size
	sp.writeUint64(filestat+40, uint64(stat.ModTime().UnixNano())) // atim
	sp.writeUint64(filestat+48, uint64(stat.ModTime().UnixNano())) // mtim
	sp.writeUint64(filestat+56, uint64(stat.ModTime().UnixNano())) // ctim

	return ErrnoSuccess
}

// PathFilestatSetTimes sets the times for a path
func (sp *WasiSnapshotPreview1) PathFilestatSetTimes(fd int32, flags int32, pathPtr int32, pathLen int32, atim int64, mtim int64, fstFlags int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	pathBytes := sp.readBytes(pathPtr, pathLen)
	fullPath := basePath + "/" + string(pathBytes)

	atime := time.Unix(0, atim)
	mtime := time.Unix(0, mtim)

	if err := os.Chtimes(fullPath, atime, mtime); err != nil {
		return ErrnoIo
	}

	return ErrnoSuccess
}

// PathLink creates a hard link
func (sp *WasiSnapshotPreview1) PathLink(oldFd int32, oldFlags int32, oldPath int32, oldPathLen int32, newFd int32, newPath int32, newPathLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	oldBasePath, oldOk := sp.preopens[oldFd]
	newBasePath, newOk := sp.preopens[newFd]

	if !oldOk || !newOk {
		return ErrnoBadf
	}

	oldPathBytes := sp.readBytes(oldPath, oldPathLen)
	newPathBytes := sp.readBytes(newPath, newPathLen)

	fullOldPath := oldBasePath + "/" + string(oldPathBytes)
	fullNewPath := newBasePath + "/" + string(newPathBytes)

	if err := os.Link(fullOldPath, fullNewPath); err != nil {
		return ErrnoIo
	}

	return ErrnoSuccess
}

// PathOpen opens a file or directory
func (sp *WasiSnapshotPreview1) PathOpen(fd int32, dirflags int32, pathPtr int32, pathLen int32, oflags int32, fsRightsBase int64, fsRightsInheriting int64, fdflags int32, openedFd int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]
	if !ok {
		return ErrnoBadf
	}

	pathBytes := sp.readBytes(pathPtr, pathLen)
	fullPath := basePath + "/" + string(pathBytes)

	var flags int
	// oflags: creat (0x1), directory (0x2), excl (0x4), trunc (0x8)
	if oflags&0x1 != 0 {
		flags |= os.O_CREATE
	}
	if oflags&0x4 != 0 {
		flags |= os.O_EXCL
	}
	if oflags&0x8 != 0 {
		flags |= os.O_TRUNC
	}

	// Check rights for read/write
	// For simplicity, open as read-write
	flags |= os.O_RDWR

	file, err := os.OpenFile(fullPath, flags, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrnoNoent
		}
		if os.IsPermission(err) {
			return ErrnoAccess
		}
		// Try read-only
		file, err = os.OpenFile(fullPath, os.O_RDONLY, 0)
		if err != nil {
			return ErrnoIo
		}
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return ErrnoIo
	}

	var filetype uint8 = FiletypeRegularFile
	if stat.IsDir() {
		filetype = FiletypeDirectory
	}

	newFd := sp.nextFd
	sp.nextFd++

	sp.openFiles[newFd] = &FileDescriptor{
		File:     file,
		Path:     fullPath,
		Filetype: filetype,
		Offset:   0,
	}

	sp.writeUint32(openedFd, uint32(newFd))
	return ErrnoSuccess
}

// PathReadLink reads the contents of a symbolic link
func (sp *WasiSnapshotPreview1) PathReadLink(fd int32, pathPtr int32, pathLen int32, buf int32, bufLen int32, bufUsed int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	pathBytes := sp.readBytes(pathPtr, pathLen)
	fullPath := basePath + "/" + string(pathBytes)

	link, err := os.Readlink(fullPath)
	if err != nil {
		return ErrnoIo
	}

	linkBytes := []byte(link)
	if int32(len(linkBytes)) > bufLen {
		linkBytes = linkBytes[:bufLen]
	}

	sp.writeBytes(buf, linkBytes)
	sp.writeUint32(bufUsed, uint32(len(linkBytes)))

	return ErrnoSuccess
}

// PathRemoveDirectory removes a directory
func (sp *WasiSnapshotPreview1) PathRemoveDirectory(fd int32, pathPtr int32, pathLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	pathBytes := sp.readBytes(pathPtr, pathLen)
	fullPath := basePath + "/" + string(pathBytes)

	if err := os.Remove(fullPath); err != nil {
		return ErrnoIo
	}

	return ErrnoSuccess
}

// PathRename renames a file or directory
func (sp *WasiSnapshotPreview1) PathRename(oldFd int32, oldPathPtr int32, oldPathLen int32, newFd int32, newPathPtr int32, newPathLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	oldBasePath, oldOk := sp.preopens[oldFd]
	newBasePath, newOk := sp.preopens[newFd]

	if !oldOk || !newOk {
		return ErrnoBadf
	}

	oldPathBytes := sp.readBytes(oldPathPtr, oldPathLen)
	newPathBytes := sp.readBytes(newPathPtr, newPathLen)

	fullOldPath := oldBasePath + "/" + string(oldPathBytes)
	fullNewPath := newBasePath + "/" + string(newPathBytes)

	if err := os.Rename(fullOldPath, fullNewPath); err != nil {
		return ErrnoIo
	}

	return ErrnoSuccess
}

// PathSymlink creates a symbolic link
func (sp *WasiSnapshotPreview1) PathSymlink(oldPath int32, oldPathLen int32, fd int32, newPath int32, newPathLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	oldPathBytes := sp.readBytes(oldPath, oldPathLen)
	newPathBytes := sp.readBytes(newPath, newPathLen)

	fullNewPath := basePath + "/" + string(newPathBytes)

	if err := os.Symlink(string(oldPathBytes), fullNewPath); err != nil {
		return ErrnoIo
	}

	return ErrnoSuccess
}

// PathUnlinkFile removes a file
func (sp *WasiSnapshotPreview1) PathUnlinkFile(fd int32, pathPtr int32, pathLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	basePath, ok := sp.preopens[fd]

	if !ok {
		return ErrnoBadf
	}

	pathBytes := sp.readBytes(pathPtr, pathLen)
	fullPath := basePath + "/" + string(pathBytes)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrnoNoent
		}
		return ErrnoIo
	}

	if stat.IsDir() {
		return ErrnoIsdir
	}

	if err := os.Remove(fullPath); err != nil {
		return ErrnoIo
	}

	return ErrnoSuccess
}

// ProcExit terminates the process with an exit code
func (sp *WasiSnapshotPreview1) ProcExit(code int32) {
	sp.exitCode = code
	sp.hasExited = true

	// In a real implementation, this would terminate the WASM execution
	fmt.Printf("WASI proc_exit called with code: %d\n", code)
	os.Exit(int(code))
}

// RandomGet fills a buffer with random bytes
func (sp *WasiSnapshotPreview1) RandomGet(buf int32, bufLen int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	randomBytes := make([]byte, bufLen)
	if _, err := rand.Read(randomBytes); err != nil {
		return ErrnoIo
	}

	sp.writeBytes(buf, randomBytes)
	return ErrnoSuccess
}

// FdReadDir reads directory entries from a directory fd
func (sp *WasiSnapshotPreview1) FdReadDir(fd int32, buf int32, bufLen int32, cookie int64, bufUsed int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	var path string
	if p, ok := sp.preopens[fd]; ok {
		path = p
	} else if fileDesc, ok := sp.openFiles[fd]; ok {
		path = fileDesc.Path
	} else {
		return ErrnoBadf
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return ErrnoIo
	}

	offset := int32(0)
	for i := int64(0); i < int64(len(entries)); i++ {
		if i < cookie {
			continue
		}

		entry := entries[i]
		name := entry.Name()

		// dirent structure:
		// - d_next: u64 (cookie of next entry)
		// - d_ino: u64
		// - d_namlen: u32
		// - d_type: u8
		entrySize := int32(24 + len(name))

		if offset+entrySize > bufLen {
			break
		}

		var filetype uint8 = FiletypeRegularFile
		if entry.IsDir() {
			filetype = FiletypeDirectory
		}

		sp.writeUint64(buf+offset, uint64(i+1))          // d_next
		sp.writeUint64(buf+offset+8, 0)                  // d_ino
		sp.writeUint32(buf+offset+16, uint32(len(name))) // d_namlen
		sp.writeBytes(buf+offset+20, []byte{filetype})   // d_type
		sp.writeBytes(buf+offset+24, []byte(name))       // name

		offset += entrySize
	}

	sp.writeUint32(bufUsed, uint32(offset))
	return ErrnoSuccess
}

// FdSync synchronizes file data to disk
func (sp *WasiSnapshotPreview1) FdSync(fd int32) int32 {

	if fileDesc, ok := sp.openFiles[fd]; ok && fileDesc.File != nil {
		if err := fileDesc.File.Sync(); err != nil {
			return ErrnoIo
		}
		return ErrnoSuccess
	}

	return ErrnoBadf
}

// PollOneOf polls for events (stub implementation)
func (sp *WasiSnapshotPreview1) PollOneOf(in int32, out int32, nsubscriptions int32, nevents int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	// This is a simplified implementation that returns immediately
	// A full implementation would actually poll for I/O events
	sp.writeUint32(nevents, 0)
	return ErrnoSuccess
}

// EnvironGet writes environment variable pointers and strings to memory
func (sp *WasiSnapshotPreview1) EnvironGet(environ int32, environBuf int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	bufOffset := environBuf
	for i, env := range sp.env {
		// Write pointer to this environment variable
		sp.writeUint32(environ+int32(i*4), uint32(bufOffset))

		// Write the environment variable string with null terminator
		envBytes := append([]byte(env), 0)
		sp.writeBytes(bufOffset, envBytes)
		bufOffset += int32(len(envBytes))
	}

	return ErrnoSuccess
}

// EnvironSizesGet returns the number of environment variables and total buffer size
func (sp *WasiSnapshotPreview1) EnvironSizesGet(environCount int32, environBufSize int32) int32 {
	if sp.memory == nil {
		return ErrnoFault
	}

	sp.writeUint32(environCount, uint32(len(sp.env)))

	totalSize := 0
	for _, env := range sp.env {
		totalSize += len(env) + 1 // +1 for null terminator
	}
	sp.writeUint32(environBufSize, uint32(totalSize))

	return ErrnoSuccess
}

func (sp *WasiSnapshotPreview1) Register(linker *wasp.Linker) error {
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
			return err
		}
	}
	return nil
}
