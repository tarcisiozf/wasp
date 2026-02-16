package memory

type DataSegment struct {
	MemoryIndex int
	Offset      int
	Data        []byte
}
