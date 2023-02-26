package golottie

import (
	"fmt"

	"github.com/chromedp/chromedp"
)

type Renderer struct {
	framesTotal int
	FramesDone  int
	ctx         Context
}

func New(ctx Context) (renderer *Renderer) {
	return &Renderer{ctx: ctx}
}

func (r *Renderer) SetAnimation(animation Animation) error {
	r.framesTotal = animation.GetFramesTotal()
	if err := chromedp.Run(r.ctx,
		chromedp.Navigate(animation.GetURL()),
	); err != nil {
		return err
	}
	return nil
}

func (r *Renderer) NextFrame() bool {
	if r.FramesDone >= r.framesTotal {
		r.ctx.Error(EOF)
		return false
	}
	if err := chromedp.Run(r.ctx,
		chromedp.Evaluate(fmt.Sprintf("anim.goToAndStop(%d, true)", r.FramesDone), nil),
	); err != nil {
		r.ctx.Error(err)
		return false
	}
	r.FramesDone++
	return true
}

func (r *Renderer) RenderFrame(frameBuf *[]byte) error {
	return chromedp.Run(r.ctx,
		chromedp.Screenshot("#lottie", frameBuf, chromedp.ByQuery))
}

func (r *Renderer) RenderFrameSVG(frameBuf *string) error {
	return chromedp.Run(r.ctx,
		chromedp.OuterHTML("svg", frameBuf, chromedp.ByQuery))
}

// func (r *Renderer) RenderAll(frames *[]*[]byte) error {
// 	return chromedp.Run(r.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
// 		var err error
// 		for i, frame := range *frames {
// 			if err = chromedp.Screenshot("#lottie", frame, chromedp.ByQuery).Do(ctx); err != nil {
// 				return err
// 			}
// 			if err = chromedp.Evaluate(fmt.Sprintf("anim.goToAndStop(%d, true)", i), nil).Do(ctx); err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	}))
// }
