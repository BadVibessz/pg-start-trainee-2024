package script

import "errors"

var (
	ErrNoSuchRunningScript    = errors.New("no such running script")
	ErrCannotCastToCancelFunc = errors.New("cannot cast cache value to context.CancelFunc")

	ErrNoSuchScript = errors.New("no such script")
)
