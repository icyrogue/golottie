package golottie

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
)

type animation struct {
	width       int
	height      int
	framesTotal int
	server      *httptest.Server

	buf      *bytes.Buffer
	Template *template.Template
}

//go:embed templates/default.gohtml
var defTemplate embed.FS

func NewAnimation(data []byte) *animation {
	return &animation{
		buf: bytes.NewBuffer(data),
	}
}

func (a *animation) WithDefaultTemplate() (animation *animation, err error) {
	if a.buf == nil || a.buf.Len() <= 0 {
		return a, fmt.Errorf("error creating new animation: %w", ErrNilAnimationData)
	}
	data := map[string]template.JS{
		"data": template.JS(a.buf.Bytes()),
	}
	a.Template, err = template.ParseFS(defTemplate, "templates/default.gohtml")
	if err != nil {
		return a, err
	}
	a.buf.Reset()
	return a, a.Template.Execute(a.buf, data)
}

func (a *animation) WithCustomTemplate(templ *template.Template) error {
	//TODO
	return nil
}

func (a *animation) GetURL() (url string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.Copy(w, a.buf)
	})
	ts := httptest.NewServer(handler)
	a.server = ts

	return ts.URL
}

func (a *animation) GetWidth() int {
	return a.width
}

func (a *animation) GetHeight() int {
	return a.height
}

func (a *animation) GetFramesTotal() int {
	return a.framesTotal
}

func (a *animation) Close() {
	if a.server != nil {
		a.server.Close()
	}
}
