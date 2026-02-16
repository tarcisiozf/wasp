package memory

type Table struct {
	ElementType byte
	InitialSize int
	MaxSize     int
	Elements    []int // function indices
}
