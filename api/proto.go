package api

type (
	Req struct {
		Metadata map[string]string // random headers.
		Body     string            // hex encoded
	}

	Res struct {
		Code        int    // http status code the proxy should use.
		ContentType string // content type the proxy should use.
		Body        string // hex encoded body the proxy should send.
	}
)
