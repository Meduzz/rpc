package messages

type (
	MsgBuilder struct {
		body any
		code int
		msg  string
	}
)

func (m *MsgBuilder) WithBody(body any) *MsgBuilder {
	m.body = body
	return m
}

func (m *MsgBuilder) WithProblem(code int, msg string) *MsgBuilder {
	m.code = code
	m.msg = msg
	return m
}

func (m *MsgBuilder) WithCode(code int) *MsgBuilder {
	m.code = code
	return m
}
