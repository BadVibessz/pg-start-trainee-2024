package db

import "errors"

var (
	ErrCannotOpenConnection = errors.New("dbutils: cannot open database connection")
)
