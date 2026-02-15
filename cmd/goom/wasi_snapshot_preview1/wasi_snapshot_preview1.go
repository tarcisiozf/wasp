package wasi_snapshot_preview1

type WasiSnapshotPreview1 struct {
}

func (sp *WasiSnapshotPreview1) ArgsGet(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) ArgsSizeGet(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) ClockResGet(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) ClockTimeGet(int32, int64, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdClose(int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdStatGet(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdStatSetFlags(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdFilestatGet(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdFilestatSetSize(int32, int64) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdFilestatSetTimes(int32, int64, int64, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdPRead(int32, int32, int32, int64, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdPrestatGet(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdPrestatDirName(int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdPWrite(int32, int32, int32, int64, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdRead(int32, int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdSeek(int32, int64, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdWrite(int32, int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathCreateDirectory(int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathFilestatGet(int32, int32, int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathFilestatSetTimes(int32, int32, int32, int32, int64, int64, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathLink(int32, int32, int32, int32, int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathOpen(int32, int32, int32, int32, int32, int64, int64, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathReadLink(int32, int32, int32, int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathRemoveDirectory(int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathRename(int32, int32, int32, int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathSymlink(int32, int32, int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PathUnlinkFile(int32, int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) ProcExit(code int32) {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) RandomGet(int32, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdReadDir(int32, int32, int32, int64, int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) FdSync(int32) int32 {
	panic("not implemented")
}

func (sp *WasiSnapshotPreview1) PollOneOf(int32, int32, int32, int32) int32 {
	panic("not implemented")
}
