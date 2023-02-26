package golottie

import (
	"net/http"
	"net/http/httptest"
)

type mockAnimation struct {
	width       int
	height      int
	framesTotal int
	data        []byte
	ts          *httptest.Server
}

func (a *mockAnimation) GetURL() (url string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(a.data)
	})
	ts := httptest.NewServer(handler)
	a.ts = ts

	return ts.URL
}

func (a *mockAnimation) GetWidth() int {
	return a.width
}

func (a *mockAnimation) GetHeight() int {
	return a.height
}

func (a *mockAnimation) GetFramesTotal() int {
	return a.framesTotal
}

func (a *mockAnimation) Close() {
	a.ts.Close()
}