package glog

import (
	"log/slog"
	"sort"
	"sync/atomic"
)

type globalFieldSnapshot struct {
	attrs []slog.Attr
}

type globalFieldState struct {
	current atomic.Pointer[globalFieldSnapshot]
}

func newGlobalFieldState(fields map[string]any) *globalFieldState {
	state := &globalFieldState{}
	state.replace(fields)
	return state
}

func (s *globalFieldState) replace(fields map[string]any) {
	if s == nil {
		return
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	snapshot := &globalFieldSnapshot{
		attrs: make([]slog.Attr, 0, len(keys)),
	}
	for _, key := range keys {
		snapshot.attrs = append(snapshot.attrs, slog.Any(key, fields[key]))
	}
	s.current.Store(snapshot)
}

func (s *globalFieldState) load() *globalFieldSnapshot {
	if s == nil {
		return nil
	}
	return s.current.Load()
}
