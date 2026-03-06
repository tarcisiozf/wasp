package memory

type Memory interface {
	Load(offset int, size int) []byte
	Store(offset int, bytes []byte)
	Grow(delta int) bool
	NumPages() int
	PageSize() int
	MaxPages() int
	Size() int
	SizeOf() uint64
	Data() []byte
}
