package api

type (
	Req struct {
		Metadata map[string]string `json:"metadata,omitempty"` // random headers.
		Body     string            `json:"body,omitempty"`     // base64 encoded
	}

	Res struct {
		Code     int               `json:"code"`               // http status code the proxy should use.
		Metadata map[string]string `json:"metadata,omitempty"` // headers the proxy should use.
		Body     string            `json:"body,omitempty"`     // base64 encoded body the proxy should send.
	}

	ErrorDTO struct {
		Message string `json:"msg"`
	}
)
