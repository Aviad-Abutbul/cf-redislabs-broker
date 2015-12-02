package redislabs

import "errors"

var (
	ErrPlanDoesNotExist    = errors.New("plan does not exist")
	ErrServiceDoesNotExist = errors.New("service does not exist")
)
