package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/goliatone/go-errors"
	"github.com/goliatone/go-logger/glog"
)

func init() {
	glog.ColorConsoleTSFormat = "15:04:05.000"
}

func myRichErrorHandler(err error) []slog.Attr {
	var richErr *errors.Error
	if errors.As(err, &richErr) {
		var attrs []slog.Attr
		if richErr.Code != 0 {
			attrs = append(attrs, slog.Int("error_code", richErr.Code))
		}

		if richErr.TextCode != "" {
			attrs = append(attrs, slog.String("text_code", richErr.TextCode))
		}

		if richErr.Category != "" {
			attrs = append(attrs, slog.String("category", richErr.Category.String()))
		}

		if richErr.RequestID != "" {
			attrs = append(attrs, slog.String("request_id", richErr.RequestID))
		}

		if len(richErr.AllValidationErrors()) > 0 {
			attrs = append(attrs, slog.Any("validation_errors", richErr.AllValidationErrors()))
		}

		if len(richErr.Metadata) > 0 {
			attrs = append(attrs, slog.Any("metadata", richErr.Metadata))
		}
		return attrs
	}
	return nil
}

type CustomErr struct {
	msg    string
	code   int
	status string
	err    error
}

func (e *CustomErr) Error() string {
	return fmt.Sprintf("%s (code: %d, status: %s)", e.msg, e.code, e.status)
}

func (e *CustomErr) Code() int {
	return e.code
}

func (e *CustomErr) Status() string {
	return e.status
}

func (e *CustomErr) Unwrap() error {
	return e.err
}

func NewCustomError(msg string, code int, status string, err error) *CustomErr {
	return &CustomErr{
		msg:    msg,
		code:   code,
		status: status,
		err:    err,
	}
}

func main() {

	log := glog.NewLogger(
		glog.WithLoggerTypePretty(), glog.WithLevel(glog.Trace),
		glog.WithName("app"),
		glog.WithRichErrorHandler(errors.ToSlogAttributes),
	)

	log.Warn("////// start /////")

	log.Focus("cron", "db")

	log.GetLogger("db").Warn("Connected to database")
	log.GetLogger("api").Debug("Server started")
	log.GetLogger("rag").Trace("Server started")
	log.GetLogger("http").Error("Server started", "source", "main.go")

	log.GetLogger("cron").Info("And here we are...", "meta", "data")

	log.Unfocus()
	log.Debug("================")

	log.GetLogger("api").Debug("Server started")
	log.GetLogger("api").Debug("Server started")

	log.GetLogger("loggername").Error("Server started", "source", "main.go")
	log.GetLogger("http").Error("Server started", "source", "main.go")
	log.GetLogger("loggername").Error("Server started", "source", "main.go")
	log.GetLogger("http").Error("Server started", "source", "main.go")

	log.GetLogger("graphql_http_service").Info("Server started", "source", "main.go")
	log.GetLogger("rag").Trace("Server started")
	log.GetLogger("loggername").Error("Server started", "source", "main.go")
	log.GetLogger("rag_service").Error("Server started", "source", "main.go")
	log.GetLogger("loggername").Error("Server started", "source", "main.go")
	log.GetLogger("cron").Error("Server started:", "source", "main.go")
	log.GetLogger("rag_service").Error("Server started", "source", "main.go", "thij", "extra")
	log.GetLogger("api").Debug("Server started")
	log.GetLogger("rag_service").Error("Server started", "source", "main.go")
	log.GetLogger("rag").Trace("Server started")
	log.GetLogger("loggername").Error("Server started:", "source", "main.go")
	log.GetLogger("cron").Warn("Server started:", "source", "main.go")
	log.GetLogger("graphql_http_service").Trace("Server started", "source", "main.go")
	log.GetLogger("rag").Trace("Server started")
	log.GetLogger("api").Warn("Connected to database")

	log.GetLogger("auth").Info("AUTH logging here")
	log.GetLogger("auth:http").Info("AUTH HTTP child logging here")
	log.GetLogger("auth:ctrl").Info("AUTH CTRL child logging here")
	// log.GetLogger("db").Warn("Connected to database")

	log.GetLogger("http").Error("Server started", "source", "main.go", "now", time.Now())
	log.GetLogger("router").Info("Registering a route and using a long message here to see what happens", "method", "POST", "route", "/test", "name", "testing-route")
	log.GetLogger("loggername").Error("Server started", "keys", glog.Args("source", "main.go", "other", "thing", "must", "go"))
	log.GetLogger("loggername").Error("Server started", "error", errors.New("this is an error"))

	oerr := errors.New("this is an error")
	log.GetLogger("loggername").Error("Server started", oerr)

	log.GetLogger("http").Error("error running query", oerr, "query", "this hsould be an object")
	log.GetLogger("http").Error("unable to add inbox to manager", "service", "google", oerr)

	richErr := errors.New("test error here", errors.CategoryAuthz).
		WithCode(errors.CodeForbidden).
		WithMetadata(map[string]any{
			"user_id":  102,
			"app_name": "test",
		}).WithRequestID("3d8f58da-6034-4861-833a-7a222d8ed8d3")

	log.GetLogger("http").Error("testing rich errors", richErr)

	///////
	err := NewCustomError("failed to setup server", 503, "Service Unavailable", oerr)
	log.GetLogger("loggername").Fatal("Server started", err)

}
