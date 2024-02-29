package ds

type DataStruct interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}
