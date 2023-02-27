package mock

import (
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

func (a *Animation) GetURL() (url string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		//nolint:errcheck
		w.Write(a.Data)
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
