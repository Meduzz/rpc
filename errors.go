package rpc

import (
	"fmt"
)

var (
	ErrUnexpectedReply = fmt.Errorf("Message did not expect reply")
)
