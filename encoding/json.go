package encoding

import "encoding/json"

type (
	jsonCodec struct{}
)

func Json() Codec {
	return &jsonCodec{}
}

func (j *jsonCodec) Marshal(it any) ([]byte, error) {
	return json.Marshal(it)
}

func (j *jsonCodec) Unmarshal(bs []byte, to any) error {
	return json.Unmarshal(bs, to)
}
