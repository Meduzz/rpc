package rpc

import (
	"fmt"
)

var (
	ErrUnexpectedReply = fmt.Errorf("message did not expect reply")
)
