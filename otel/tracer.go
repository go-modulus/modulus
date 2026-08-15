package otel

import (
	"runtime"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns otel.Tracer with the name of the current package (import path).
func Tracer(opts ...trace.TracerOption) trace.Tracer {
	pc, _, _, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()

	if i := strings.LastIndex(name, "/"); i != -1 {
		if j := strings.Index(name[i+1:], "."); j != -1 {
			name = name[:i+1+j]
		}
	}
	return otel.Tracer(name, opts...)
}
