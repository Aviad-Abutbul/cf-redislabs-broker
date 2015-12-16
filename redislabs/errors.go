package redislabs

import "errors"

var (
	ErrPlanDoesNotExist       = errors.New("plan does not exist")
	ErrServiceDoesNotExist    = errors.New("service does not exist")
	ErrDatabaseNameIsRequired = errors.New("a database name prefix is required: specify one via the -c option")
)
