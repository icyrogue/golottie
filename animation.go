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

	"github.com/ysmood/gson"
)

// AnimationData implements Animation interface to serve animation data localy.
type AnimationData struct {
	// Template field contains pointer to animation template
	// to be reused when parsing multiple animations.
	// Equals to nil if animation template hasn't been initialized.
	Template *template.Template

	framesTotal int
	server      *httptest.Server
	buf         *bytes.Buffer
}

//go:embed templates/default.gohtml
var defTemplate embed.FS

// NewAnimation creates a new animation with data argument as animation data.
// Template function should be called on resulting animation for it to be
// completely initialized.
//
// Example:
//
//	data, _ := os.ReadFile("animation.json")
//	animation, err := golottie.NewAnimation(data).WithDefaultTemplate()
//	if err != nil {
//		log.Fatal(err)
//	}
//	renderer.SetAnimation(animation)
func NewAnimation(data []byte) *AnimationData {
	return &AnimationData{
		buf:         bytes.NewBuffer(data),
		framesTotal: gson.New(data).Get("op").Int(),
	}
}

// WithDefaultTemplate initializes animation data using embeded default template.
// Returns an error if the initial data is nil or has 0 length.
func (a *AnimationData) WithDefaultTemplate() (animation *AnimationData, err error) {
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

// WithCustomTemplate initializes animation data using provided custom template.
// Can be called to reuse animation.Template multiple times.
//
// Example:
//
//	data, _ := os.ReadFile("animation.json")
//	animation, _ := golottie.NewAnimation(data).WithDefaultTemplate()
//	data, _ = os.ReadFile("animation2.json")
//	animationTwo, err := golottie.NewAnimation(data).WithCustomTemplate(animation.Template, nil)
//	if err != nil {
//		log.Fatal(err.Error)
//	}
//
// The data arguments map can be provided to be available inside the template.
// If data map is nil, a new map will be created with "animationData" key containing
// initial animation data as unescaped JS string. If data map isn't nil, "animationData"
// key will be added to it.
func (a *AnimationData) WithCustomTemplate(templ *template.Template, data map[string]interface{}) (_ *AnimationData, err error) {
	if templ == nil {
		return a, fmt.Errorf("Error parsing custom template: %w", ErrNilTemplate)
	}
	if data == nil {
		data = make(map[string]interface{})
	}
	data["animationData"] = template.JS(a.buf.Bytes())
	return a, templ.Execute(a.buf, data)
}

// GetURL serves an animation data localy and returns an URL to be used by renderer.
func (a *AnimationData) GetURL() (url string) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		//nolint:errcheck // in case of an error renderer will get nothing and error out
		io.Copy(w, a.buf)
	})
	ts := httptest.NewServer(handler)
	a.server = ts

	return ts.URL
}

// GetFramesTotal is used by renderer to get the amount of frames to render.
func (a *AnimationData) GetFramesTotal() int {
	return a.framesTotal
}

// Close closes the local server if it exists.
func (a *AnimationData) Close() {
	if a.server != nil {
		a.server.Close()
	}
}
