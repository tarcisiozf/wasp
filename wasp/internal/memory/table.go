package memory

type Table struct {
	ElementType byte
	InitialSize int
	MaxSize     int
	Elements    []int // function indices
}

func (table Table) Clone() *Table {
	elements := make([]int, len(table.Elements))
	copy(elements, table.Elements)

	return &Table{
		ElementType: table.ElementType,
		InitialSize: table.InitialSize,
		MaxSize:     table.MaxSize,
		Elements:    elements,
	}
}
