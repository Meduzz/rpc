package encoding

type (
	Codec interface {
		Marshal(any) ([]byte, error)
		Unmarshal([]byte, any) error
	}
)
