package golottie

import (
	"context"
	"errors"
)

var (
	EOF                 = errors.New("EOF")
	ErrNilAnimationData = errors.New("animation data is nil")
)

type Context interface {
	context.Context
	Error(err error)
}

type Animation interface {
	Close()
	GetURL() string

	GetFramesTotal() int
	GetWidth() int
	GetHeight() int
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
