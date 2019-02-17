package rpc

import (
	"errors"
	"fmt"
)

var (
	ErrTimeout         = errors.New("Timeout was reached")
	ErrUnexpectedReply = fmt.Errorf("Message did not expect reply")
)
