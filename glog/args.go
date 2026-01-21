package glog

import (
	"fmt"
	"log/slog"
)

func argsToAttrSlice(args []any) []any {
	var (
		attr  slog.Attr
		attrs []any
	)
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}

	return attrs
}

const badKey = "!BADKEY"

func argsToAttr(args []any) (slog.Attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, fmt.Sprintf("missing value for key %q", x)), nil
		}
		return slog.Any(x, args[1]), args[2:]

	case slog.Attr:
		return x, args[1:]

	default:
		return slog.String(badKey, fmt.Sprintf("expected key string, got %T (%v)", x, x)), args[1:]
	}
}
