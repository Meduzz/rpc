package api

type (
	Req struct {
		Metadata map[string]string
		Body     string
	}

	Res struct {
		Code        int
		ContentType string
		Body        string
	}
)
