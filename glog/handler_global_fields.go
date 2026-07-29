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

	recordAttrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		recordAttrs = append(recordAttrs, attr)
		return true
	})
	recordAttrs, changed := normalizeAttrs(recordAttrs)

	explicitKeys := cloneKeySet(h.boundKeys)
	collectTopLevelAttrKeys(explicitKeys, recordAttrs)

	attrs := make([]slog.Attr, 0, len(snapshot.attrs))
	for _, attr := range snapshot.attrs {
		if _, exists := explicitKeys[attr.Key]; exists {
			continue
		}
		attrs = append(attrs, attr)
	}
	if !changed && len(attrs) == 0 {
		return h.next.Handle(ctx, record)
	}

	if changed {
		resolved := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
		resolved.AddAttrs(recordAttrs...)
		resolved.AddAttrs(attrs...)
		return h.next.Handle(ctx, resolved)
	}

	record = record.Clone()
	record.AddAttrs(attrs...)
	return h.next.Handle(ctx, record)
}

func (h *globalFieldsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	attrs, _ = normalizeAttrs(attrs)
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

// normalizeAttrs resolves LogValuer values and removes empty attributes and
// groups before collision analysis. The normalized attributes are also passed
// to the wrapped handler so each LogValuer is evaluated at most once.
func normalizeAttrs(attrs []slog.Attr) ([]slog.Attr, bool) {
	var normalized []slog.Attr
	changed := false

	for index, attr := range attrs {
		attr, visible, attrChanged := normalizeAttr(attr)
		if (attrChanged || !visible) && !changed {
			normalized = make([]slog.Attr, 0, len(attrs))
			normalized = append(normalized, attrs[:index]...)
			changed = true
		}
		if !visible {
			continue
		}
		if changed {
			normalized = append(normalized, attr)
		}
	}

	if !changed {
		return attrs, false
	}
	return normalized, true
}

func normalizeAttr(attr slog.Attr) (slog.Attr, bool, bool) {
	if attr.Equal(slog.Attr{}) {
		return slog.Attr{}, false, true
	}

	value := attr.Value
	changed := false
	if value.Kind() == slog.KindLogValuer {
		value = value.Resolve()
		attr.Value = value
		changed = true
	}
	if attr.Equal(slog.Attr{}) {
		return slog.Attr{}, false, true
	}

	if value.Kind() != slog.KindGroup {
		return attr, true, changed
	}

	group, groupChanged := normalizeAttrs(value.Group())
	if len(group) == 0 {
		return slog.Attr{}, false, true
	}
	if changed || groupChanged {
		attr.Value = slog.GroupValue(group...)
		changed = true
	}
	return attr, true, changed
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
