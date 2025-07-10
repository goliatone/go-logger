package glog

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgsToAttr(t *testing.T) {
	t.Run("string key and value", func(t *testing.T) {
		args := []any{"key", "value", "more"}
		attr, remaining := argsToAttr(args)
		assert.Equal(t, slog.Any("key", "value"), attr)
		assert.Equal(t, []any{"more"}, remaining)
	})

	t.Run("slog.Attr", func(t *testing.T) {
		slogAttr := slog.Int("id", 123)
		args := []any{slogAttr, "more"}
		attr, remaining := argsToAttr(args)
		assert.Equal(t, slogAttr, attr)
		assert.Equal(t, []any{"more"}, remaining)
	})

	t.Run("lone string becomes bad key", func(t *testing.T) {
		args := []any{"a lonely string"}
		attr, remaining := argsToAttr(args)
		assert.Equal(t, slog.String(badKey, "a lonely string"), attr)
		assert.Nil(t, remaining)
	})

	t.Run("non-string key becomes bad key", func(t *testing.T) {
		args := []any{123, "value"}
		attr, remaining := argsToAttr(args)
		assert.Equal(t, slog.Any(badKey, 123), attr)
		assert.Equal(t, []any{"value"}, remaining)
	})
}

func TestArgsToAttrSlice(t *testing.T) {
	t.Run("valid pairs", func(t *testing.T) {
		args := []any{"key1", "val1", "key2", 2}
		attrs := argsToAttrSlice(args)
		expected := []any{
			slog.Any("key1", "val1"),
			slog.Any("key2", 2),
		}
		assert.Equal(t, expected, attrs)
	})

	t.Run("mixed with slog.Attr", func(t *testing.T) {
		slogAttr := slog.Bool("active", true)
		args := []any{"key1", "val1", slogAttr}
		attrs := argsToAttrSlice(args)
		expected := []any{
			slog.Any("key1", "val1"),
			slogAttr,
		}
		assert.Equal(t, expected, attrs)
	})

	t.Run("trailing key", func(t *testing.T) {
		args := []any{"key1", "val1", "trailing_key"}
		attrs := argsToAttrSlice(args)
		expected := []any{
			slog.Any("key1", "val1"),
			slog.String(badKey, "trailing_key"),
		}
		assert.Equal(t, expected, attrs)
	})

	t.Run("empty args", func(t *testing.T) {
		args := []any{}
		attrs := argsToAttrSlice(args)
		assert.Empty(t, attrs)
	})
}
