package integration

import (
	"errors"
)

type Ignore func(err error) bool

func Combine(errs ...error) Ignore {
	return func(err error) bool {
		for _, checkErr := range errs {
			if errors.As(err, &checkErr) {
				return true
			}
		}
		return false
	}
}
