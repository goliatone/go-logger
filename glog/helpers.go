package glog

import "context"

var nopLoggerInstance Logger = &nopLogger{}

type nopLogger struct{}

func (n *nopLogger) Trace(msg string, args ...any) {}

func (n *nopLogger) Debug(msg string, args ...any) {}

func (n *nopLogger) Info(msg string, args ...any) {}

func (n *nopLogger) Warn(msg string, args ...any) {}

func (n *nopLogger) Error(msg string, args ...any) {}

func (n *nopLogger) Fatal(msg string, args ...any) {}

func (n *nopLogger) WithContext(ctx context.Context) Logger {
	return n
}

type loggerProvider struct {
	logger Logger
}

func (p *loggerProvider) GetLogger(name string) Logger {
	return p.logger
}

type fallbackProvider struct {
	provider LoggerProvider
	fallback Logger
}

func (p *fallbackProvider) GetLogger(name string) Logger {
	if p.provider == nil {
		return p.fallback
	}

	logger := p.provider.GetLogger(name)
	if logger != nil {
		return logger
	}

	return p.fallback
}

// Nop returns a canonical no-op logger.
func Nop() Logger {
	return nopLoggerInstance
}

// Ensure guarantees a non-nil Logger.
func Ensure(logger Logger) Logger {
	if logger == nil {
		return Nop()
	}
	return logger
}

// ProviderFromLogger wraps a single logger as a provider.
func ProviderFromLogger(logger Logger) LoggerProvider {
	return &loggerProvider{
		logger: Ensure(logger),
	}
}

// ProviderWithFallback wraps a provider and applies a fallback when nil is returned.
func ProviderWithFallback(provider LoggerProvider, fallback Logger) LoggerProvider {
	return &fallbackProvider{
		provider: provider,
		fallback: Ensure(fallback),
	}
}

// Resolve resolves provider and logger with precedence provider > logger > nop.
func Resolve(name string, provider LoggerProvider, logger Logger) (LoggerProvider, Logger) {
	if provider == nil {
		resolvedLogger := Ensure(logger)
		return ProviderFromLogger(resolvedLogger), resolvedLogger
	}

	resolvedProvider := ProviderWithFallback(provider, logger)
	return resolvedProvider, resolvedProvider.GetLogger(name)
}
