package pucci

import "errors"

// Errors
var (
	ErrorTableDoesNotExist = errors.New("table does not exist")
	ErrorItemDoesNotExist  = errors.New("item does not exist")
	ErrorDuplicatedID      = errors.New("duplicated id")
	ErrorNilDocument       = errors.New("received document is nil")
)
