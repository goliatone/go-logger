package glog

import (
	"context"
	"log/slog"
)

// globalFieldsHandler adds the current root-owned field snapshot to each
// record. Explicit attributes bound to the logger or supplied with the record
// take precedence over global defaults at the current group level.
type globalFieldsHandler struct {
	next      slog.Handler
	fields    *globalFieldState
	boundKeys map[string]struct{}
}

func newGlobalFieldsHandler(next slog.Handler, fields *globalFieldState) slog.Handler {
	return &globalFieldsHandler{
		next:      next,
		fields:    fields,
		boundKeys: map[string]struct{}{},
	}
}

func (h *globalFieldsHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *globalFieldsHandler) Handle(ctx context.Context, record slog.Record) error {
	snapshot := h.fields.load()
	if snapshot == nil || len(snapshot.attrs) == 0 {
		return h.next.Handle(ctx, record)
	}

	explicitKeys := cloneKeySet(h.boundKeys)
	record.Attrs(func(attr slog.Attr) bool {
		collectTopLevelAttrKeys(explicitKeys, []slog.Attr{attr})
		return true
	})

	attrs := make([]slog.Attr, 0, len(snapshot.attrs))
	for _, attr := range snapshot.attrs {
		if _, exists := explicitKeys[attr.Key]; exists {
			continue
		}
		attrs = append(attrs, attr)
	}
	if len(attrs) == 0 {
		return h.next.Handle(ctx, record)
	}

	cloned := record.Clone()
	cloned.AddAttrs(attrs...)
	return h.next.Handle(ctx, cloned)
}

func (h *globalFieldsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	boundKeys := cloneKeySet(h.boundKeys)
	collectTopLevelAttrKeys(boundKeys, attrs)
	return &globalFieldsHandler{
		next:      h.next.WithAttrs(attrs),
		fields:    h.fields,
		boundKeys: boundKeys,
	}
}

func (h *globalFieldsHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &globalFieldsHandler{
		next:      h.next.WithGroup(name),
		fields:    h.fields,
		boundKeys: map[string]struct{}{},
	}
}

func cloneKeySet(source map[string]struct{}) map[string]struct{} {
	cloned := make(map[string]struct{}, len(source))
	for key := range source {
		cloned[key] = struct{}{}
	}
	return cloned
}

func collectTopLevelAttrKeys(keys map[string]struct{}, attrs []slog.Attr) {
	for _, attr := range attrs {
		if attr.Equal(slog.Attr{}) {
			continue
		}

		value := attr.Value
		if value.Kind() != slog.KindGroup {
			keys[attr.Key] = struct{}{}
			continue
		}

		group := value.Group()
		if !hasVisibleAttrs(group) {
			continue
		}
		if attr.Key == "" {
			collectTopLevelAttrKeys(keys, group)
			continue
		}
		keys[attr.Key] = struct{}{}
	}
}

func hasVisibleAttrs(attrs []slog.Attr) bool {
	for _, attr := range attrs {
		if attr.Equal(slog.Attr{}) {
			continue
		}
		value := attr.Value
		if value.Kind() != slog.KindGroup || hasVisibleAttrs(value.Group()) {
			return true
		}
	}
	return false
}
