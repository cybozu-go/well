package cmd

import (
	"context"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	t.Parallel()

	ctx := WithRequestID(context.Background(), "abc")
	if v := ctx.Value(RequestIDContextKey); v == nil {
		t.Error(`v == nil`)
	} else {
		if v != "abc" {
			t.Error(`v != "abc"`)
		}
	}
}

func TestBackgroundWithID(t *testing.T) {
	t.Parallel()

	ctx := WithRequestID(context.Background(), "abc")
	ctx = BackgroundWithID(ctx)
	if v := ctx.Value(RequestIDContextKey); v == nil {
		t.Error(`v == nil`)
	} else {
		if v != "abc" {
			t.Error(`v != "abc"`)
		}
	}
}
