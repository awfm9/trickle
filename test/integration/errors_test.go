package integration

import (
	"errors"
)

var (
	errFinished = errors.New("test finished")
	errTimeout  = errors.New("test timed out")
)
