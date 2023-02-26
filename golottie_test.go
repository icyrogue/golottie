package golottie

import (
	"context"
	_ "embed"
	"errors"
	"testing"
	"time"

	"github.com/icyrogue/golottie/internal/mock"
	"github.com/stretchr/testify/assert"
)

//go:embed misc/test.html
var mockHTML []byte

//go:embed misc/test_badHTML.html
var badHTML []byte

// noAnimHTML contains no animation data which is necessary
// to move to the next frame
var noAnimHTML = []byte("<html><body>( ´･_･`)</body></html>")

var (
	okAnimation = mock.Animation{
		Width:       600,
		Height:      600,
		FramesTotal: 68,
		Data:        mockHTML,
	}
	badHTMLAnimation = mock.Animation{
		Width:       600,
		Height:      600,
		FramesTotal: 68,
		Data:        badHTML,
	}
	zeroFramesAnimation = mock.Animation{
		Width:       600,
		Height:      600,
		FramesTotal: 0,
		Data:        mockHTML,
	}
	noAnimDataAnimation = mock.Animation{
		Width:       600,
		Height:      600,
		FramesTotal: 68,
		Data:        noAnimHTML,
	}
)

func Test_RenderFrame(t *testing.T) {
	p, c := context.WithTimeout(context.Background(), 5*time.Second)
	defer c()
	ctx, cancel := NewContext(p)
	defer cancel()
	renderer := New(ctx)

	// Since we are reusing the context and its error buf, we should keep track
	// of how many errors were in it before the particular test
	prevErrorLen := len(ctx.Errors)

	tests := []struct {
		name        string
		animation   animation
		expectedErr error
	}{
		{
			name:        "OK_animation",
			animation:   &okAnimation,
			expectedErr: nil,
		},
		{
			name:        "ZeroFrames_animation",
			animation:   &zeroFramesAnimation,
			expectedErr: EOF,
		},
		{
			name:        "NoData_animation",
			animation:   &noAnimDataAnimation,
			expectedErr: errors.New("Exeption"),
		},
		{
			name:        "BadHTML_animation",
			animation:   &badHTMLAnimation,
			expectedErr: errors.New("context canceled"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			// Update the animation to render
			err := renderer.SetAnimation(tt.animation)
			assert.NoError(t, err)
			// Go to the first frame and check the error buf, it will always
			// return false and push to ctx errors if something went wrong
			if !renderer.NextFrame() {
				err = ctx.Errors[prevErrorLen]
				prevErrorLen++
				// Some errors in chromedp are not pre-defined so for now,
				// just check if error exists or doesn't
				if tt.expectedErr != nil {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				return
			}

			// Render frame into designated buffer
			var buf []byte
			err = renderer.RenderFrame(&buf)

			// Some errors in chromedp are not pre-defined so for now,
			// just check if error exists or doesn't
			if tt.expectedErr != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			// Check if frame has been actually rendered and err buf is clean
			assert.Equal(t, prevErrorLen, len(ctx.Errors),
				"render context has errors\nwhen it shouldn't:\n%s", ctx.Errors)
			assert.Greater(t, len(buf), 0)
		})
	}
}

func Test_RenderFrameSVG(t *testing.T) {
	p, c := context.WithTimeout(context.Background(), 2*time.Second)
	defer c()
	ctx, cancel := NewContext(p)
	defer cancel()
	renderer := New(ctx)

	// Since we are reusing the context and its error buf, we should keep track
	// of how many errors were in it before the particular test
	prevErrorLen := len(ctx.Errors)

	tests := []struct {
		name        string
		animation   animation
		expectedErr error
	}{
		{
			name:        "OK_animation",
			animation:   &okAnimation,
			expectedErr: nil,
		},
		{
			name:        "BadHTML_animation",
			animation:   &badHTMLAnimation,
			expectedErr: errors.New("context canceled"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			// Update the animation to render
			err := renderer.SetAnimation(tt.animation)
			assert.NoError(t, err)
			// Go to the first frame and check the error buf, it will always
			// return false and push to ctx errors if something went wrong
			assert.True(t, renderer.NextFrame())

			// Render frame into designated buffer
			var buf string
			err = renderer.RenderFrameSVG(&buf)

			// Some errors in chromedp are not pre-defined so for now,
			// just check if error exists or doesn't
			if tt.expectedErr != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			// Check if frame has been actually rendered and err buf is clean
			assert.Equal(t, prevErrorLen, len(ctx.Errors),
				"render context has errors\nwhen it shouldn't:\n%s", ctx.Errors)
			assert.Greater(t, len(buf), 0)
		})
	}
}

type noURL struct {
	mock.Animation
}

func (u *noURL) GetURL(_ Context) string {
	return "(ノಠ益ಠ)ノ彡┻━┻"
}

func Test_InvalidAnimation(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()
	renderer := New(ctx)
	assert.Error(t, renderer.SetAnimation(&noURL{}))
}
