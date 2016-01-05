package instancecreators

import "errors"

var (
	ErrFailedToLoadState            = errors.New("failed to load the broker state")
	ErrInstanceExists               = errors.New("such instance already exists")
	ErrFailedToSaveState            = errors.New("failed to save the new broker state")
	ErrFailedToCreateDatabase       = errors.New("failed to create a database")
	ErrCreateDatabaseTimeoutExpired = errors.New("create database timeout expired")
)
