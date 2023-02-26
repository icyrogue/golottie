package golottie

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_error(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	defer cancel()
	err := errors.New("OI!")
	ctx.Error(err)
	assert.ErrorIs(t, ctx.Errors[0], err)
}
