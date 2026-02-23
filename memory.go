package wasp

type Memory interface {
	Load(offset int, size int) []byte
	Store(offset int, bytes []byte)
}
