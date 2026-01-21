# go-logger

A flexible, structured logging library for Go built on top of Go's standard `slog` package. It provides colored console output, multiple log levels including trace, contextual logging, and a focus feature for filtering logs during development.

## Features

- **Multiple Log Levels**: Support for TRACE, DEBUG, INFO, WARN, ERROR, and FATAL levels
- **Multiple Output Formats**: JSON, Console (text), and Pretty (colored) output formats
- **Colored Console Output**: Beautiful, human-readable logs with color-coded levels for development
- **Named Loggers**: Create and manage multiple named logger instances
- **Focus Mode**: Filter logs to show only specific logger names during debugging
- **Contextual Logging**: Attach context and structured attributes to log entries
- **Error Enrichment**: Automatic extraction of error metadata including stack traces
- **Thread-Safe**: Safe for concurrent use across goroutines
- **Customizable**: Flexible configuration through functional options

## Installation

```bash
go get github.com/goliatone/go-logger
```

## Quick Start

```go
package main

import (
    "github.com/goliatone/go-logger/glog"
)

func main() {
    // Create a logger with pretty (colored) output
    log := glog.NewLogger(
        glog.WithLoggerTypePretty(),
        glog.WithLevel(glog.Info),
    )

    // Basic logging
    log.Info("Application started")
    log.Debug("Debug information", "user_id", 123)
    log.Warn("Warning message", "retry_count", 3)
    log.Error("Error occurred", "error", err)
}
```

## Configuration Options

### Creating a Logger

The logger can be configured using functional options:

```go
log := glog.NewLogger(
    glog.WithName("myapp"),              // Set logger name
    glog.WithLevel(glog.Debug),          // Set log level
    glog.WithLoggerTypePretty(),         // Use colored console output
    glog.WithAddSource(true),            // Include source file info
    glog.WithContext(ctx),               // Attach context
    glog.WithRichErrorHandler(handler),  // Custom error attribute extraction
    glog.WithWriter(os.Stdout),          // Override output writer
    glog.WithExitFunc(func(int) {}),     // Override Fatal exit behavior
)
```

### Available Options

- `WithName(string)` - Set the logger name
- `WithLevel(string)` - Set the minimum log level (TRACE, DEBUG, INFO, WARN, ERROR)
- `WithLoggerType(string)` - Set output format (json, console, pretty)
- `WithLoggerTypeJSON()` - Use JSON output format
- `WithLoggerTypeConsole()` - Use plain text output format
- `WithLoggerTypePretty()` - Use colored console output format
- `WithAddSource(bool)` - Include source file and line number in logs
- `WithContext(context.Context)` - Attach a context to the logger
- `WithRichErrorHandler(RichErrorHandler)` - Set custom error attribute extractor
- `WithHandlerWrapper(func(slog.Handler) slog.Handler)` - Wrap the base slog handler before focus/name handling
- `WithWriter(io.Writer)` - Override the output writer (e.g. multi-writer for rotation)
- `WithExitFunc(func(int))` - Override the exit behavior used by `Fatal`

## Log Levels

The logger supports the following levels (from lowest to highest severity):

- `TRACE` - Most detailed debugging information
- `DEBUG` - Debugging information
- `INFO` - General informational messages
- `WARN` - Warning messages
- `ERROR` - Error messages with stack traces
- `FATAL` - Fatal errors that terminate the program

```go
log.Trace("Detailed trace info")
log.Debug("Debug information")
log.Info("General information")
log.Warn("Warning message")
log.Error("Error occurred", errorInstance)
log.Fatal("Fatal error", err) // This will exit the program by default
```

Use `WithExitFunc` to customize or disable the exit behavior during testing or in libraries.

## Named Loggers

Create child loggers with specific names for different components:

```go
mainLog := glog.NewLogger(glog.WithLoggerTypePretty())

// Create named loggers for different components
dbLog := mainLog.GetLogger("database")
apiLog := mainLog.GetLogger("api")
authLog := mainLog.GetLogger("auth")

dbLog.Info("Connected to database")
apiLog.Debug("API request received", "method", "GET", "path", "/users")
authLog.Warn("Invalid token attempted")
```

## Focus Mode

Focus mode allows you to filter logs to show only specific logger names during development:

```go
log := glog.NewLogger(glog.WithLoggerTypePretty())

// Create multiple named loggers
dbLog := log.GetLogger("db")
apiLog := log.GetLogger("api")
cronLog := log.GetLogger("cron")

// Focus on specific loggers (only these will output)
log.Focus("db", "cron")

dbLog.Info("This will be shown")     // ✓ Shown
apiLog.Info("This will be hidden")   // ✗ Hidden
cronLog.Info("This will be shown")   // ✓ Shown

// Remove focus to show all logs again
log.Unfocus()
```

## Structured Logging

Add structured data to your logs using key-value pairs:

```go
// Using alternating key-value pairs
log.Info("User logged in",
    "user_id", 12345,
    "ip", "192.168.1.1",
    "timestamp", time.Now(),
)

// Using the Args helper for multiple attributes
log.Debug("Request processed", glog.Args(
    "method", "POST",
    "path", "/api/users",
    "duration_ms", 234,
    "status", 200,
))

// Using With to create a logger with persistent fields
requestLog := log.With(
    "request_id", "abc-123",
    "user_id", 456,
)
requestLog.Info("Processing started")
requestLog.Info("Processing completed")
```

Malformed key/value pairs are captured under the `!BADKEY` attribute with a descriptive message.

### Optional Structured Fields

Loggers that implement `glog.FieldsLogger` allow you to attach a map of fields in a single call while keeping the base `Logger` interface minimal:

```go
if fl, ok := log.(glog.FieldsLogger); ok {
    log = fl.WithFields(map[string]any{
        "request_id": reqID,
        "customer_id": customerID,
    })
}

log.Info("Request received")
```

## Error Handling

The logger provides enhanced error handling with automatic stack trace capture:

```go
// Simple error logging - automatically captures stack trace
err := doSomething()
log.Error("Operation failed", err)

// Error with additional context
log.Error("Database query failed",
    err,
    "query", "SELECT * FROM users",
    "duration_ms", 1234,
)

// Custom error types implementing Code() or Status() interfaces
type MyError struct {
    msg string
    code int
}

func (e *MyError) Error() string { return e.msg }
func (e *MyError) Code() int { return e.code }

myErr := &MyError{msg: "validation failed", code: 400}
log.Error("Request failed", myErr) // Will include error_code: 400

// Fatal errors (exits program with error code)
log.Fatal("Critical failure", err)
```

### Rich Error Handler

You can provide a custom error handler to extract additional attributes from errors:

```go
func myErrorHandler(err error) []slog.Attr {
    var attrs []slog.Attr

    // Extract custom error fields
    if customErr, ok := err.(*MyCustomError); ok {
        attrs = append(attrs,
            slog.String("error_type", customErr.Type),
            slog.Int("error_code", customErr.Code),
        )
    }

    return attrs
}

log := glog.NewLogger(
    glog.WithRichErrorHandler(myErrorHandler),
)
```

## Output Formats

### JSON Format

```go
log := glog.NewLogger(glog.WithLoggerTypeJSON())
log.Info("User action", "user_id", 123, "action", "login")
// Output: {"ts":"2024-01-01T12:00:00Z","level":"info","msg":"User action","user_id":123,"action":"login"}
```

### Console Format

```go
log := glog.NewLogger(glog.WithLoggerTypeConsole())
log.Info("User action", "user_id", 123)
// Output: ts=2024-01-01T12:00:00Z level=info msg="User action" user_id=123
```

### Pretty Format (Colored)

```go
log := glog.NewLogger(glog.WithLoggerTypePretty())
log.Info("User action", "user_id", 123)
// Output: Colored output with timestamp, level, and formatted attributes
```

## Context Support

Attach context to loggers for request tracing:

```go
// Create a logger with context
ctx := context.WithValue(context.Background(), "request_id", "xyz-789")
ctxLog := log.WithContext(ctx)

// All logs from ctxLog will include the context
ctxLog.Info("Processing request")
```

## Advanced Usage

### Wrapping the Handler

Wrap the base slog handler to inject extra attributes or integrate with otel-style handlers:

```go
log := glog.NewLogger(
    glog.WithLoggerTypeJSON(),
    glog.WithHandlerWrapper(func(h slog.Handler) slog.Handler {
        return otelwrap.WrapHandler(h)
    }),
)
```

### Custom Output Writer

Override the output writer to support file rotation or tee output:

```go
writer := io.MultiWriter(os.Stdout, lumberjackLogger)
log := glog.NewLogger(
    glog.WithLoggerTypeJSON(),
    glog.WithWriter(writer),
)
```

### Customizing Timestamp Format

For colored console output, you can customize the timestamp format:

```go
func init() {
    glog.ColorConsoleTSFormat = "15:04:05.000" // Just time
    // or
    glog.ColorConsoleTSFormat = "2006-01-02 15:04:05" // Date and time
}
```

### Thread-Safe Logger Sharing

The logger is thread-safe and can be shared across goroutines:

```go
log := glog.NewLogger(glog.WithLoggerTypePretty())

var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        log.Info("Goroutine processing", "id", id)
    }(i)
}
wg.Wait()
```

## Examples

### Complete Example

```go
package main

import (
    "context"
    "errors"
    "github.com/goliatone/go-logger/glog"
)

func main() {
    // Initialize logger
    log := glog.NewLogger(
        glog.WithName("app"),
        glog.WithLoggerTypePretty(),
        glog.WithLevel(glog.Debug),
    )

    // Component-specific loggers
    dbLog := log.GetLogger("database")
    apiLog := log.GetLogger("api")

    // Startup logs
    log.Info("Application starting", "version", "1.0.0")

    // Database operations
    dbLog.Debug("Connecting to database", "host", "localhost")
    dbLog.Info("Database connected successfully")

    // API operations with structured data
    apiLog.Info("Server started",
        "port", 8080,
        "environment", "production",
    )

    // Error handling
    err := errors.New("connection timeout")
    dbLog.Error("Query failed", err, "query", "SELECT * FROM users")

    // Focus on specific components for debugging
    log.Focus("database")

    dbLog.Debug("This will be shown")
    apiLog.Debug("This will be hidden due to focus")

    log.Unfocus()
}
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
