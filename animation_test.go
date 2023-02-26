package golottie

import (
	"bytes"
	"context"
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
		name string
		data []byte
		err  error
	}{
		{
			name: "OK_animation",
			data: animData,
			err:  nil,
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
			defer animation.Close()
			assert.ErrorIs(t, err, tt.err)
			if err != nil {
				t.Log("wat", err)
				return
			}
			ctx, cancel := NewContext(context.Background())
			defer cancel()
			url := animation.GetURL()
			assert.Empty(t, ctx.Errors)
			resp, err := http.Get(url)
			assert.Equal(t, resp.StatusCode, http.StatusOK)
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			testBody := parseTemplate(t, tt.data, defTemplate)
			assert.True(t, bytes.Equal(testBody, body), "Animation data is corrupted")
		})
	}
}

func parseTemplate(t *testing.T, animData []byte, templateBytes fs.FS) []byte {
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
