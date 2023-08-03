package golottie

import (
	"context"

	"github.com/chromedp/chromedp"
)

type gContext struct {
	context.Context
	Errors []error
}

// NewContext wraps a new [chromedp] context created from parent ctx.
//
// [chromedp]: https://github.com/chromedp/chromedp
func NewContext(ctx context.Context) (context *gContext, cancel context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:])
	aCtx, _ := chromedp.NewExecAllocator(ctx, opts...)
	dpContext, _ := chromedp.NewContext(aCtx)

	return &gContext{
		Context: dpContext,
	}, cancel
}

// Error pushes an error to context error stack.
func (c *gContext) Error(err error) {
	c.Errors = append(c.Errors, err)
}
