package test

import "github.com/rossmacarthur/fudge/errors"

var ErrTest = errors.Sentinel("test error", "ERR_12345", "test")
