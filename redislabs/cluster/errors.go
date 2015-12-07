package cluster

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidJSON = errors.New("invalid JSON")
)

func ErrUnknownParam(k string) error {
	return fmt.Errorf("%s property is not supported", k)
}

func ErrInvalidType(k string) error {
	return fmt.Errorf("%s value is of the wrong type", k)
}
