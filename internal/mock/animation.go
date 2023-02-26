package mock

import (
	"context"
	"net/http"
	"net/http/httptest"
)

type Animation struct {
	Width       int
	Height      int
	FramesTotal int
	Data        []byte
	ts          *httptest.Server
}

type gContext interface {
	context.Context
	Error(error)
}

func (a *Animation) GetURL(what gContext) (url string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, err := w.Write(a.Data)
		if err != nil {
			//ctx.Error(err)
		}
	})
	ts := httptest.NewServer(handler)
	a.ts = ts

	return ts.URL
}

func (a *Animation) GetWidth() int {
	return a.Width
}

func (a *Animation) GetHeight() int {
	return a.Height
}

func (a *Animation) GetFramesTotal() int {
	return a.FramesTotal
}

func (a *Animation) Close() {
	a.ts.Close()
}