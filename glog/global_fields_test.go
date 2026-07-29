package glog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalFieldStateCopiesAndSortsInput(t *testing.T) {
	fields := map[string]any{
		"service":     "orders",
		"environment": "development",
	}
	state := newGlobalFieldState(fields)
	fields["service"] = "mutated"
	fields["added"] = true

	snapshot := state.load()
	require.NotNil(t, snapshot)
	require.Len(t, snapshot.attrs, 2)
	assert.Equal(t, "environment", snapshot.attrs[0].Key)
	assert.Equal(t, "development", snapshot.attrs[0].Value.Any())
	assert.Equal(t, "service", snapshot.attrs[1].Key)
	assert.Equal(t, "orders", snapshot.attrs[1].Value.Any())
}

func TestSetGlobalFieldsReplacesRootOwnedSnapshot(t *testing.T) {
	var buf bytes.Buffer
	root := newTestLogger(
		&buf,
		WithGlobalFields(map[string]any{"service": "orders", "stale": true}),
	)
	child := root.GetLogger("child").(*BaseLogger)
	derived := child.WithFields(map[string]any{"request_id": "req-1"}).(*BaseLogger)

	require.Same(t, root.globalFields, child.globalFields)
	require.Same(t, root.globalFields, derived.globalFields)

	returned := child.SetGlobalFields(map[string]any{"service": "billing"})
	assert.Same(t, child, returned)

	snapshot := root.globalFields.load()
	require.NotNil(t, snapshot)
	require.Len(t, snapshot.attrs, 1)
	assert.Equal(t, "service", snapshot.attrs[0].Key)
	assert.Equal(t, "billing", snapshot.attrs[0].Value.Any())

	root.SetGlobalFields(nil)
	require.Empty(t, root.globalFields.load().attrs)
}

func TestGlobalFieldsPropagateReplaceAndClear(t *testing.T) {
	var buf bytes.Buffer
	root := newTestLogger(
		&buf,
		WithLoggerTypeJSON(),
		WithGlobalFields(map[string]any{
			"environment": "development",
			"service":     "orders",
		}),
	)
	existingChild := root.GetLogger("existing").(*BaseLogger)
	derived := existingChild.WithContext(context.Background()).(*BaseLogger).
		With("request_id", "req-1")

	root.Info("root initial")
	existingChild.Info("child initial")
	derived.Info("derived initial")

	root.SetGlobalFields(map[string]any{
		"service": "billing",
		"version": "v2",
	})
	futureChild := root.GetLogger("future")
	root.Info("root replaced")
	existingChild.Info("child replaced")
	derived.Info("derived replaced")
	futureChild.Info("future replaced")

	root.SetGlobalFields(nil)
	existingChild.Info("child cleared")

	records := decodeJSONRecords(t, buf.String())
	require.Len(t, records, 8)
	for _, record := range records[:3] {
		assert.Equal(t, "orders", record["service"])
		assert.Equal(t, "development", record["environment"])
	}
	for _, record := range records[3:7] {
		assert.Equal(t, "billing", record["service"])
		assert.Equal(t, "v2", record["version"])
		assert.NotContains(t, record, "environment")
	}
	assert.NotContains(t, records[7], "service")
	assert.NotContains(t, records[7], "version")
}

func TestExplicitFieldsOverrideGlobalDefaults(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(
		&buf,
		WithLoggerTypeJSON(),
		WithGlobalFields(map[string]any{
			"logger":     "global-name",
			"request_id": "global-request",
			"service":    "global-service",
		}),
	)

	logger.GetLogger("api").(*BaseLogger).
		WithFields(map[string]any{"service": "local-service"}).
		Info("collision", "request_id", "local-request")

	line := strings.TrimSpace(buf.String())
	record := decodeJSONRecords(t, line)[0]
	assert.Equal(t, "api", record["logger"])
	assert.Equal(t, "local-request", record["request_id"])
	assert.Equal(t, "local-service", record["service"])
	assert.Equal(t, 1, strings.Count(line, `"logger":`))
	assert.Equal(t, 1, strings.Count(line, `"request_id":`))
	assert.Equal(t, 1, strings.Count(line, `"service":`))
}

func TestHandlerWrapperObservesGlobalAndEnrichedFieldsOnce(t *testing.T) {
	sink := &recordCaptureSink{}
	logger := NewLogger(
		WithWriter(io.Discard),
		WithGlobalFields(map[string]any{"service": "orders"}),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sink, next: next}
		}),
	)

	logger.Error("request failed", fmt.Errorf("outer: %w", errors.New("root cause")))

	records := sink.snapshot(t)
	require.Len(t, records, 1)
	attrs := recordAttrs(records[0].record)
	assert.Equal(t, "orders", attrs["service"])
	assert.Equal(t, "root cause", attrs["root_error"])
	assert.EqualError(t, attrs["error"].(error), "outer: root cause")
	assert.NotEmpty(t, attrs["stack"])
	require.NotZero(t, records[0].record.PC)
}

func TestGlobalFieldsPreserveGroupsAndGroupCollisions(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(
		&buf,
		WithLoggerTypeJSON(),
		WithGlobalFields(map[string]any{
			"environment": "development",
			"service":     "global-service",
		}),
	)

	handler := logger.logger.Handler().
		WithGroup("request").
		WithAttrs([]slog.Attr{slog.String("service", "local-service")})
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "grouped", 0)
	record.AddAttrs(slog.String("request_id", "req-1"))
	require.NoError(t, handler.Handle(context.Background(), record))

	decoded := decodeJSONRecords(t, buf.String())[0]
	group, ok := decoded["request"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "development", group["environment"])
	assert.Equal(t, "local-service", group["service"])
	assert.Equal(t, "req-1", group["request_id"])
	assert.Equal(t, 1, strings.Count(buf.String(), `"service":`))
}

func TestGlobalFieldsSurviveHandlerRebuilds(t *testing.T) {
	var buf bytes.Buffer
	root := newTestLogger(&buf, WithLoggerTypeJSON(), WithLevel(Info))
	child := root.GetLogger("child").(*BaseLogger)
	derived := child.With("scope", "derived")

	root.SetGlobalFields(map[string]any{"service": "orders"})
	root.WithLevel(Debug)
	root.WithLoggerType(" json ")
	root.Focus("child")
	child.Debug("child rebuilt")
	derived.Info("derived existing")
	root.Unfocus()

	records := decodeJSONRecords(t, buf.String())
	require.Len(t, records, 2)
	for _, record := range records {
		assert.Equal(t, "orders", record["service"])
	}
}

func TestGlobalFieldsHandleInlineGroupsAndIgnoreEmptyGroups(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(
		&buf,
		WithLoggerTypeJSON(),
		WithGlobalFields(map[string]any{
			"service": "global-service",
			"version": "v1",
		}),
	)

	logger.Info(
		"inline",
		slog.Group("", slog.String("service", "local-service")),
		slog.Group("version"),
	)

	line := strings.TrimSpace(buf.String())
	record := decodeJSONRecords(t, line)[0]
	assert.Equal(t, "local-service", record["service"])
	assert.Equal(t, "v1", record["version"])
	assert.Equal(t, 1, strings.Count(line, `"service":`))
	assert.Equal(t, 1, strings.Count(line, `"version":`))
}

type countingLogValuer struct {
	calls *int
	value string
}

func (v countingLogValuer) LogValue() slog.Value {
	(*v.calls)++
	return slog.StringValue(v.value)
}

func TestGlobalFieldCollisionScanDoesNotResolveValuesEarly(t *testing.T) {
	var buf bytes.Buffer
	calls := 0
	logger := newTestLogger(
		&buf,
		WithLoggerTypeJSON(),
		WithGlobalFields(map[string]any{"service": "global-service"}),
	)

	logger.Info("log valuer", "service", countingLogValuer{
		calls: &calls,
		value: "local-service",
	})

	record := decodeJSONRecords(t, buf.String())[0]
	assert.Equal(t, "local-service", record["service"])
	assert.Equal(t, 1, calls)
}

func TestGlobalFieldsAreIndependentAcrossRoots(t *testing.T) {
	sinkA := &recordCaptureSink{}
	sinkB := &recordCaptureSink{}
	rootA := NewLogger(
		WithWriter(io.Discard),
		WithGlobalFields(map[string]any{"service": "a"}),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sinkA, next: next}
		}),
	)
	rootB := NewLogger(
		WithWriter(io.Discard),
		WithGlobalFields(map[string]any{"service": "b"}),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sinkB, next: next}
		}),
	)

	rootA.SetGlobalFields(map[string]any{"service": "a2"})
	rootA.Info("root a")
	rootB.Info("root b")

	recordsA := sinkA.snapshot(t)
	recordsB := sinkB.snapshot(t)
	require.Len(t, recordsA, 1)
	require.Len(t, recordsB, 1)
	assert.Equal(t, "a2", recordAttrs(recordsA[0].record)["service"])
	assert.Equal(t, "b", recordAttrs(recordsB[0].record)["service"])
}

func TestConcurrentGlobalFieldReplacementPublishesCompleteSnapshots(t *testing.T) {
	sink := &recordCaptureSink{}
	logger := NewLogger(
		WithWriter(io.Discard),
		WithGlobalFields(map[string]any{
			"generation": "a",
			"only_a":     true,
		}),
		WithHandlerWrapper(func(next slog.Handler) slog.Handler {
			return &recordCaptureHandler{sink: sink, next: next}
		}),
	)

	const (
		loggers    = 12
		recordsPer = 100
		replaces   = 500
	)
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(loggers + 1)
	for worker := range loggers {
		go func() {
			defer wg.Done()
			<-start
			child := logger.GetLogger(fmt.Sprintf("worker-%d", worker))
			for sequence := range recordsPer {
				child.Info("concurrent", "sequence", sequence)
			}
		}()
	}
	go func() {
		defer wg.Done()
		<-start
		for i := range replaces {
			if i%2 == 0 {
				logger.SetGlobalFields(map[string]any{
					"generation": "a",
					"only_a":     true,
				})
				continue
			}
			logger.SetGlobalFields(map[string]any{
				"generation": "b",
				"only_b":     true,
			})
		}
	}()
	close(start)
	wg.Wait()

	records := sink.snapshot(t)
	require.Len(t, records, loggers*recordsPer)
	for _, captured := range records {
		attrs := recordAttrs(captured.record)
		switch attrs["generation"] {
		case "a":
			assert.Equal(t, true, attrs["only_a"])
			assert.NotContains(t, attrs, "only_b")
		case "b":
			assert.Equal(t, true, attrs["only_b"])
			assert.NotContains(t, attrs, "only_a")
		default:
			t.Fatalf("record observed incomplete snapshot: %#v", attrs)
		}
	}
}

func decodeJSONRecords(t *testing.T, output string) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	records := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var record map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &record))
		records = append(records, record)
	}
	return records
}
