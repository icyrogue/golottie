package golottie

import (
	"bytes"
	"embed"
	_ "embed"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed misc/test.json
var animData []byte

func Test_NewAnimation(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		err    error
		frames int
	}{
		{
			name:   "OK_animation",
			data:   animData,
			err:    nil,
			frames: 68,
		},
		{
			name: "Nil_animation",
			data: nil,
			err:  ErrNilAnimationData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			animation, err := NewAnimation(tt.data).WithDefaultTemplate()
			//nolint:all // Animation.close() doesn't return an error to check
			defer animation.Close()
			assert.ErrorIs(t, err, tt.err)
			if err != nil {
				t.Log("wat", err)
				return
			}
			url := animation.GetURL()
			resp, err := http.Get(url)
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, http.StatusOK)
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			testBody := parseTemplate(t, tt.data, defTemplate)
			assert.True(t, bytes.Equal(testBody, body), "Animation data is corrupted")
			assert.Equal(t, tt.frames, animation.GetFramesTotal(), "animation total frame count doesn't match")
		})
	}
	t.Run("NilDefTemplate_animation", func(t *testing.T) {
		animation := NewAnimation(animData)
		defer animation.Close()
		copy := defTemplate
		defTemplate = embed.FS{}
		_, err := animation.WithDefaultTemplate()
		defTemplate = copy
		assert.Error(t, err)
	})
}

func Test_WithCustomTemplate(t *testing.T) {

	tests := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "OK_data",
			data: map[string]interface{}{"customData": "༼⍢༽"},
		},
		{
			name: "Nil_data",
			data: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte("༼∵༽༼⍨༽")
			templ, err := template.New("test").Parse("{{ .data}}{{ .customData}}")
			assert.NoError(t, err)
			a, err := NewAnimation(data).WithCustomTemplate(templ, tt.data)
			assert.NoError(t, err)
			cd, ok := tt.data["customData"].(string)
			if !ok {
				cd = ""
			}
			testData := append(data, cd...)
			assert.Truef(t, bytes.Equal(a.buf.Bytes(), testData), "Expected: %s\nGot: %s", testData, a.buf.Bytes())
		})
	}
	t.Run("Nil_template", func(t *testing.T) {
		_, err := NewAnimation(nil).WithCustomTemplate(nil, nil)
		assert.ErrorIs(t, err, ErrNilTemplate)
	})
}

func parseTemplate(t *testing.T, animData []byte, templateFS fs.FS) []byte {
	data := map[string]template.JS{
		"data": template.JS(animData),
	}
	var buf bytes.Buffer
	err := template.Must(template.ParseFS(defTemplate, "templates/default.gohtml")).Execute(&buf, data)
	if err != nil {
		t.Fatal(err.Error())
	}
	return buf.Bytes()
}
