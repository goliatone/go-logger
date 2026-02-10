package glog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testLogger struct {
	name string
}

func (l *testLogger) Trace(msg string, args ...any) {}

func (l *testLogger) Debug(msg string, args ...any) {}

func (l *testLogger) Info(msg string, args ...any) {}

func (l *testLogger) Warn(msg string, args ...any) {}

func (l *testLogger) Error(msg string, args ...any) {}

func (l *testLogger) Fatal(msg string, args ...any) {}

func (l *testLogger) WithContext(ctx context.Context) Logger {
	return l
}

type fixedProvider struct {
	logger Logger
}

func (p *fixedProvider) GetLogger(name string) Logger {
	return p.logger
}

type nilProvider struct{}

func (p *nilProvider) GetLogger(name string) Logger {
	return nil
}

func TestNopLoggerIsNilSafe(t *testing.T) {
	logger := Nop()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() { logger.Trace("trace") })
	assert.NotPanics(t, func() { logger.Debug("debug") })
	assert.NotPanics(t, func() { logger.Info("info") })
	assert.NotPanics(t, func() { logger.Warn("warn") })
	assert.NotPanics(t, func() { logger.Error("error") })
	assert.NotPanics(t, func() { logger.Fatal("fatal") })

	withCtx := logger.WithContext(context.Background())
	assert.NotNil(t, withCtx)
	assert.Same(t, logger, withCtx)
}

func TestEnsure(t *testing.T) {
	assert.Same(t, Nop(), Ensure(nil))

	logger := &testLogger{name: "custom"}
	assert.Same(t, logger, Ensure(logger))
}

func TestProviderFromLoggerUsesFallbackForNil(t *testing.T) {
	provider := ProviderFromLogger(nil)
	assert.NotNil(t, provider.GetLogger("admin"))

	logger := &testLogger{name: "provided"}
	provider = ProviderFromLogger(logger)
	assert.Same(t, logger, provider.GetLogger("admin"))
}

func TestProviderWithFallback(t *testing.T) {
	fallback := &testLogger{name: "fallback"}
	provider := ProviderWithFallback(nil, fallback)
	assert.Same(t, fallback, provider.GetLogger("admin"))

	primary := &testLogger{name: "primary"}
	provider = ProviderWithFallback(&fixedProvider{logger: primary}, fallback)
	assert.Same(t, primary, provider.GetLogger("admin"))

	provider = ProviderWithFallback(&nilProvider{}, fallback)
	assert.Same(t, fallback, provider.GetLogger("admin"))
}

func TestResolvePrecedenceAndNilSafety(t *testing.T) {
	providerLogger := &testLogger{name: "provider"}
	logger := &testLogger{name: "logger"}

	resolvedProvider, resolvedLogger := Resolve("admin", &fixedProvider{logger: providerLogger}, logger)
	assert.Same(t, providerLogger, resolvedLogger)
	assert.Same(t, providerLogger, resolvedProvider.GetLogger("admin.child"))

	resolvedProvider, resolvedLogger = Resolve("admin", &nilProvider{}, logger)
	assert.Same(t, logger, resolvedLogger)
	assert.Same(t, logger, resolvedProvider.GetLogger("admin.child"))

	resolvedProvider, resolvedLogger = Resolve("admin", nil, logger)
	assert.Same(t, logger, resolvedLogger)
	assert.Same(t, logger, resolvedProvider.GetLogger("admin.child"))

	resolvedProvider, resolvedLogger = Resolve("admin", nil, nil)
	assert.NotNil(t, resolvedLogger)
	assert.NotNil(t, resolvedProvider.GetLogger("admin.child"))
	assert.NotPanics(t, func() {
		resolvedLogger.WithContext(context.Background()).Info("safe")
	})
}
