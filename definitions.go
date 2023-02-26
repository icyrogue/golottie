package golottie

import (
	"context"
	"errors"
)

var (
	EOF = errors.New("EOF")
)

type Context interface {
	context.Context
	Error(err error)
}

type Frame struct {
	Buf    *string
	Num    int
	Width  int
	Height int
}

var (
	ServeError = errors.New("local server hasn't been initialized")
)
