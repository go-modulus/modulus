package errlog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"braces.dev/errtrace"
	"github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errlog"
	"github.com/go-modulus/modulus/errors/errsys"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/stretchr/testify/require"
)

type logRow struct {
	Time  time.Time   `json:"time"`
	Level string      `json:"level"`
	Msg   string      `json:"msg"`
	Error logRowError `json:"error"`
}

type logRowError struct {
	Type    string            `json:"type"`
	Message string            `json:"message"`
	Hint    string            `json:"hint"`
	Trace   []string          `json:"trace"`
	Meta    map[string]string `json:"meta,omitempty"`
	Cause   *logRowError      `json:"cause,omitempty"`
}

func TestLogError(t *testing.T) {
	t.Parallel()
	t.Run(
		"log native error", func(t *testing.T) {
			t.Parallel()
			err := fmt.Errorf("error")
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("	Then the error should be logged")
			require.Equal(t, "error", row.Error.Message)
			require.Equal(t, "*errors.errorString", row.Error.Type)
			require.Equal(t, "", row.Error.Hint)
			require.Equal(t, 0, len(row.Error.Trace))
			require.Nil(t, row.Error.Cause)
			require.Equal(t, "error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
		},
	)

	t.Run(
		"log errsys error", func(t *testing.T) {
			t.Parallel()
			err := errsys.New("error", "hint")
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("	Then the error should be logged")
			require.Equal(t, "error", row.Error.Message)
			require.Equal(t, "errors.mError", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 0, len(row.Error.Trace))
			require.Nil(t, row.Error.Cause)
			require.Equal(t, "error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
		},
	)

	t.Run(
		"log user error", func(t *testing.T) {
			t.Parallel()
			err := erruser.New("error", "hint")
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelInfo,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("	Then the error should be logged")
			require.Equal(t, "error", row.Error.Message)
			require.Equal(t, "errors.mError", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 0, len(row.Error.Trace))
			require.Nil(t, row.Error.Cause)
			require.Equal(t, "error", row.Msg)
			require.Equal(t, "INFO", row.Level)
		},
	)

	t.Run(
		"log errsys error with meta", func(t *testing.T) {
			t.Parallel()
			err := errors.WithMeta(errsys.New("error", "hint"), "key", "value")
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("	Then the error should be logged")
			require.Equal(t, "error", row.Error.Message)
			require.Equal(t, "errors.mError", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 0, len(row.Error.Trace))
			require.Nil(t, row.Error.Cause)
			require.Equal(t, "error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
			require.Equal(t, map[string]string{"key": "value"}, row.Error.Meta)
		},
	)

	t.Run(
		"log errsys error with meta and cause", func(t *testing.T) {
			t.Parallel()
			cause := fmt.Errorf("cause")
			err := errsys.WithCause(errors.WithMeta(errsys.New("error", "hint"), "key", "value"), cause)
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("\tThen the error should be logged")
			require.Equal(t, "error", row.Error.Message)
			require.Equal(t, "errors.mError", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 0, len(row.Error.Trace))
			require.Equal(t, "error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
			require.Equal(t, map[string]string{"key": "value"}, row.Error.Meta)

			t.Log("\tAnd the cause should be logged")
			require.Equal(t, "cause", row.Error.Cause.Message)
			require.Equal(t, "*errors.errorString", row.Error.Cause.Type)
			require.Equal(t, "", row.Error.Cause.Hint)
			require.Equal(t, 0, len(row.Error.Cause.Trace))
			require.Nil(t, row.Error.Cause.Cause)
		},
	)

	t.Run(
		"log errsys error with errtrace trace and cause", func(t *testing.T) {
			t.Parallel()
			cause := errtrace.Wrap(fmt.Errorf("cause"))
			err := errtrace.Wrap(errsys.WithCause(errsys.New("error", "hint"), cause))
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("\tThen the error should be logged")
			require.Equal(t, "error", row.Error.Message)
			require.Equal(t, "*errtrace.errTrace", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 3, len(row.Error.Trace))
			require.Contains(t, row.Error.Trace[0], "github.com/go-modulus/modulus/errors/errlog_test.TestLogError")
			require.Equal(t, "error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
			require.Nil(t, row.Error.Meta)

			t.Log("\tAnd the cause should be logged")
			require.Equal(t, "cause", row.Error.Cause.Message)
			require.Equal(t, "*errtrace.errTrace", row.Error.Cause.Type)
			require.Equal(t, "", row.Error.Cause.Hint)
			require.Equal(t, 2, len(row.Error.Cause.Trace))
			require.Contains(
				t,
				row.Error.Cause.Trace[0],
				"github.com/go-modulus/modulus/errors/errlog_test.TestLogError",
			)
			require.Nil(t, row.Error.Cause.Cause)
		},
	)

	t.Run(
		"log errsys error with trace and cause", func(t *testing.T) {
			t.Parallel()
			cause := errors.WithTrace(fmt.Errorf("cause"))
			err := errors.WithTrace(errsys.WithCause(errsys.New("error", "hint"), cause))
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("\tThen the error should be logged")
			require.Equal(t, "error", row.Error.Message)
			require.Equal(t, "errors.mError", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 2, len(row.Error.Trace))
			require.Contains(t, row.Error.Trace[0], "modulus/errors/errlog/log_test.go")
			require.Equal(t, "error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
			require.Nil(t, row.Error.Meta)

			t.Log("\tAnd the cause should be logged")
			require.Equal(
				t,
				"cause",
				row.Error.Cause.Message,
				"the internal-error wrapper should be skipped when logging the cause",
			)
			require.Equal(t, "*errors.errorString", row.Error.Cause.Type)
			require.Equal(t, "", row.Error.Cause.Hint)
			require.Equal(t, 0, len(row.Error.Cause.Trace))
			require.Nil(t, row.Error.Cause.Cause)
		},
	)

	t.Run(
		"log native error WithHint call", func(t *testing.T) {
			t.Parallel()
			err := errors.WithHint(fmt.Errorf("cause"), "hint")
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("\tThen the error should be logged")
			require.Equal(t, "internal-error", row.Error.Message)
			require.Equal(t, "errors.mError", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 1, len(row.Error.Trace))
			require.Contains(t, row.Error.Trace[0], "modulus/errors/errlog/log_test.go")
			require.Equal(t, "internal-error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
			require.Nil(t, row.Error.Meta)

			t.Log("\tAnd the cause should be logged")
			require.Equal(
				t,
				"cause",
				row.Error.Cause.Message,
				"the internal-error wrapper should be skipped when logging the cause",
			)
			require.Equal(t, "*errors.errorString", row.Error.Cause.Type)
			require.Equal(t, "", row.Error.Cause.Hint)
			require.Equal(t, 0, len(row.Error.Cause.Trace))
			require.Nil(t, row.Error.Cause.Cause)
		},
	)

	t.Run(
		"log native error with WithHint and WithTrace calls", func(t *testing.T) {
			t.Parallel()
			err := errors.WithHint(fmt.Errorf("cause"), "hint")
			err = errors.WithTrace(err)
			var buf bytes.Buffer
			logger := slog.New(
				slog.NewJSONHandler(
					&buf, &slog.HandlerOptions{
						Level: slog.LevelDebug,
					},
				),
			)
			_, logged := errlog.LogError(
				t.Context(),
				err,
				logger,
				slog.LevelError,
			)

			logStr := buf.String()
			fmt.Println(logStr)
			var row *logRow
			parseErr := json.Unmarshal([]byte(logStr), &row)

			t.Log("When log an error")
			require.True(t, logged)
			require.NoError(t, parseErr)
			t.Log("\tThen the error should be logged")
			require.Equal(t, "internal-error", row.Error.Message)
			require.Equal(t, "errors.mError", row.Error.Type)
			require.Equal(t, "hint", row.Error.Hint)
			require.Equal(t, 2, len(row.Error.Trace))
			require.Contains(t, row.Error.Trace[0], "modulus/errors/errlog/log_test.go")
			require.Equal(t, "internal-error", row.Msg)
			require.Equal(t, "ERROR", row.Level)
			require.Nil(t, row.Error.Meta)

			t.Log("\tAnd the cause should be logged")
			require.Equal(
				t,
				"cause",
				row.Error.Cause.Message,
				"the internal-error wrapper should be skipped when logging the cause",
			)
			require.Equal(t, "*errors.errorString", row.Error.Cause.Type)
			require.Equal(t, "", row.Error.Cause.Hint)
			require.Equal(t, 0, len(row.Error.Cause.Trace))
			require.Nil(t, row.Error.Cause.Cause)
		},
	)
}
