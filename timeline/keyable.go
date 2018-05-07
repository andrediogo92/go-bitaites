package timeline

type Keyable interface {
	Key() []byte
	AsBinary() ([]byte, error)
	FromBinary([]byte) (error)
}
