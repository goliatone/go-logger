package glog

import (
	"fmt"
	"log/slog"
)

func normalizeArgs(args []any) []any {
	if len(args) == 0 {
		return args
	}

	out := make([]any, 0, len(args))
	expectingValue := false
	normalizeArgsInto(&out, args, &expectingValue)
	return out
}

func normalizeArgsInto(out *[]any, args []any, expectingValue *bool) {
	for _, arg := range args {
		switch v := arg.(type) {
		case ArgsList:
			if *expectingValue {
				*out = append(*out, v)
				*expectingValue = false
			} else {
				normalizeArgsInto(out, []any(v), expectingValue)
			}
			continue
		}

		*out = append(*out, arg)

		if *expectingValue {
			*expectingValue = false
			continue
		}

		switch arg.(type) {
		case slog.Attr:
		case string:
			*expectingValue = true
		}
	}
}

func argsToAttrSlice(args []any) []any {
	args = normalizeArgs(args)

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
		remaining := args[2:]
		if len(remaining) == 0 {
			remaining = nil
		}
		return slog.Any(x, args[1]), remaining

	case slog.Attr:
		remaining := args[1:]
		if len(remaining) == 0 {
			remaining = nil
		}
		return x, remaining

	default:
		remaining := args[1:]
		if len(remaining) == 0 {
			remaining = nil
		}
		return slog.String(badKey, fmt.Sprintf("expected key string, got %T (%v)", x, x)), remaining
	}
}
