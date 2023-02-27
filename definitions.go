package golottie

import (
	"context"
	"errors"
)

var (
	EOF                 = errors.New("EOF")
	ErrNilAnimationData = errors.New("animation data is nil")
	ErrNilTemplate      = errors.New("custom template is nil")
)

// Context interface is a custom context which implements context.Context
// and adds custom error method to contain an error stack.
type Context interface {
	context.Context
	Error(err error)
}

// Animation interface is used by renderer to get animation data.
type Animation interface {
	// GetURL is called by renderer to retrive animation HTML.
	GetURL() string
	// Close is needed in a case of animation data being served localy.
	Close()
	// GetFramesTotal returns number of frames to be rendered.
	GetFramesTotal() int
}
