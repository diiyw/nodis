package nodis

import "errors"

var (
	ErrClosed          = errors.New("nodis: database closed")
	ErrNotFound        = errors.New("nodis: resource not found")
	ErrAlreadyExists   = errors.New("nodis: resource already exists")
	ErrPermission      = errors.New("nodis: permission denied")
	ErrConflict        = errors.New("nodis: object revision conflict")
	ErrInvalidValue    = errors.New("nodis: unsupported or invalid value")
	ErrInvalidName     = errors.New("nodis: invalid name")
	ErrInvalidState    = errors.New("nodis: invalid transaction state")
	ErrCorruptData     = errors.New("nodis: corrupt data")
	ErrPipelineMissing = errors.New("nodis: pipeline not registered")
	ErrInvalidQuery    = errors.New("nodis: invalid query")
	ErrQueryTooLarge   = errors.New("nodis: projected object exceeds query page byte limit")
)
