package golottie

import (
	"context"

	"github.com/chromedp/chromedp"
)

type gContext struct {
	context.Context
	Errors []error
}

func NewContext(ctx context.Context) (context *gContext, cancel context.CancelFunc) {
	dpContext, cancel := chromedp.NewContext(ctx)

	return &gContext{
		Context: dpContext,
	}, cancel
}

func (c *gContext) Error(err error) {
	c.Errors = append(c.Errors, err)
}
