package redislabs

import "errors"

var (
	ErrPlanDoesNotExist        = errors.New("plan does not exist")
	ErrInstanceCreatorNotFound = errors.New("instance creator not found")
)
