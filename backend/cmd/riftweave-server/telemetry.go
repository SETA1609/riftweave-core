// Telemetry for the HTTP API (BE_S12_02, docs/backend/observability.md): OpenTelemetry traces,
// metrics, and logs, here in the adapter only, so the core stays free of any telemetry package.
// Off unless OTEL_EXPORTER_OTLP_ENDPOINT is set (OTLP/HTTP) or --telemetry names a file; a server
// with neither writes only its slog JSON lines. Every record names a command, a status, and an
// error kind, never a player's input.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// serviceName is the server's service.name (docs/backend/observability.md#resource-who-is-reporting).
const serviceName = "riftweave-server"

// The attributes this adapter writes.
const (
	attributeCommand      = "riftweave.command.name"
	attributeErrorType    = "error.type"
	attributeRules        = "riftweave.rules_version"
	attributeStatus       = "http.response.status_code"
	attributeServiceName  = "service.name"
	attributeServiceBuild = "service.version"
	metricCommandDuration = "riftweave.command.duration"
	spanCommandPrefix     = "riftweave.command "
)

// The environment variables the server follows (OpenTelemetry's).
const (
	environmentEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
	environmentDisabled = "OTEL_SDK_DISABLED"
)

// telemetryModeKind is where the records go.
type telemetryModeKind uint8

const (
	telemetryOff    telemetryModeKind = iota // only the slog JSON lines
	telemetryFile                            // the stdout exporters, into the --telemetry file
	telemetryExport                          // OTLP/HTTP to OTEL_EXPORTER_OTLP_ENDPOINT
)

// telemetryMode picks where the records go: OTEL_SDK_DISABLED wins, then an endpoint, then a file.
func telemetryMode(getenv func(string) string, file string) telemetryModeKind {
	if strings.EqualFold(getenv(environmentDisabled), "true") {
		return telemetryOff
	}
	if getenv(environmentEndpoint) != "" {
		return telemetryExport
	}
	if file != "" {
		return telemetryFile
	}
	return telemetryOff
}

// telemetryOptions are the exporters, the line output, and the build; a nil exporter records
// nothing of its signal. batch sends through batch processors, as a real exporter wants.
type telemetryOptions struct {
	spans   sdktrace.SpanExporter
	metrics sdkmetric.Reader
	logs    sdklog.Exporter
	lines   io.Writer
	build   string
	batch   bool
}

// telemetry is what the server observes with.
type telemetry struct {
	tracers    trace.TracerProvider
	meters     metric.MeterProvider
	propagator propagation.TextMapPropagator
	tracer     trace.Tracer
	duration   metric.Float64Histogram
	logger     *slog.Logger
	shutdowns  []func(context.Context) error
}

// shutdown flushes and stops every provider.
func (observer *telemetry) shutdown(ctx context.Context) error {
	var failures []error
	for _, stop := range observer.shutdowns {
		failures = append(failures, stop(ctx))
	}
	return errors.Join(failures...)
}

// newTelemetry builds the providers from the options.
func newTelemetry(ctx context.Context, options telemetryOptions) (*telemetry, error) {
	if options.lines == nil {
		options.lines = os.Stderr
	}
	if options.build == "" {
		options.build = build
	}
	described, err := resource.New(ctx,
		resource.WithAttributes(attribute.String(attributeServiceName, serviceName), attribute.String(attributeServiceBuild, options.build), attribute.String(attributeRules, core.RulesVersion)),
		resource.WithFromEnv())
	if err != nil {
		return nil, err
	}
	var observer *telemetry = &telemetry{propagator: propagation.TraceContext{}, tracers: tracenoop.NewTracerProvider(), meters: metricnoop.NewMeterProvider()}
	if options.spans != nil {
		var processor sdktrace.SpanProcessor = sdktrace.NewSimpleSpanProcessor(options.spans)
		if options.batch {
			processor = sdktrace.NewBatchSpanProcessor(options.spans)
		}
		var provider *sdktrace.TracerProvider = sdktrace.NewTracerProvider(sdktrace.WithResource(described), sdktrace.WithSpanProcessor(processor))
		observer.tracers, observer.shutdowns = provider, append(observer.shutdowns, provider.Shutdown)
	}
	if options.metrics != nil {
		var provider *sdkmetric.MeterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithResource(described), sdkmetric.WithReader(options.metrics))
		observer.meters, observer.shutdowns = provider, append(observer.shutdowns, provider.Shutdown)
		if err := runtime.Start(runtime.WithMeterProvider(provider)); err != nil {
			return nil, err
		}
	}
	var handler slog.Handler = withTrace{slog.NewJSONHandler(options.lines, nil)}
	if options.logs != nil {
		var processor sdklog.Processor = sdklog.NewSimpleProcessor(options.logs)
		if options.batch {
			processor = sdklog.NewBatchProcessor(options.logs)
		}
		var provider *sdklog.LoggerProvider = sdklog.NewLoggerProvider(sdklog.WithResource(described), sdklog.WithProcessor(processor))
		observer.shutdowns = append(observer.shutdowns, provider.Shutdown)
		handler = fanout{handler, otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(provider))}
	}
	observer.logger = slog.New(handler)
	observer.tracer = observer.tracers.Tracer(serviceName)
	observer.duration, err = observer.meters.Meter(serviceName).Float64Histogram(metricCommandDuration, metric.WithUnit("s"), metric.WithDescription("How long one core command took."))
	if err != nil {
		return nil, err
	}
	return observer, nil
}

// quietTelemetry records nothing and writes no line: the handler of a test that does not observe.
func quietTelemetry() *telemetry {
	observer, _ := newTelemetry(context.Background(), telemetryOptions{lines: io.Discard})
	return observer
}

// telemetryFor is the server's telemetry for the mode: the exporters it names, or none.
func telemetryFor(ctx context.Context, mode telemetryModeKind, file string) (*telemetry, error) {
	switch mode {
	case telemetryExport:
		spans, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, err
		}
		metrics, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, err
		}
		logs, err := otlploghttp.New(ctx)
		if err != nil {
			return nil, err
		}
		return newTelemetry(ctx, telemetryOptions{spans: spans, metrics: sdkmetric.NewPeriodicReader(metrics), logs: logs, batch: true})
	case telemetryFile:
		output, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("--telemetry %s: %w", file, err)
		}
		spans, err := stdouttrace.New(stdouttrace.WithWriter(output))
		if err != nil {
			return nil, err
		}
		metrics, err := stdoutmetric.New(stdoutmetric.WithWriter(output))
		if err != nil {
			return nil, err
		}
		logs, err := stdoutlog.New(stdoutlog.WithWriter(output))
		if err != nil {
			return nil, err
		}
		return newTelemetry(ctx, telemetryOptions{spans: spans, metrics: sdkmetric.NewPeriodicReader(metrics, sdkmetric.WithInterval(time.Minute)), logs: logs, batch: true})
	}
	return newTelemetry(ctx, telemetryOptions{})
}

// withTrace adds the trace and span ids of the record's context to a JSON line.
type withTrace struct{ slog.Handler }

func (handler withTrace) Handle(ctx context.Context, record slog.Record) error {
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		record.AddAttrs(slog.String("trace_id", span.TraceID().String()), slog.String("span_id", span.SpanID().String()))
	}
	return handler.Handler.Handle(ctx, record)
}

func (handler withTrace) WithAttrs(attributes []slog.Attr) slog.Handler {
	return withTrace{handler.Handler.WithAttrs(attributes)}
}

func (handler withTrace) WithGroup(name string) slog.Handler {
	return withTrace{handler.Handler.WithGroup(name)}
}

// fanout sends a record to every handler that takes its level.
type fanout []slog.Handler

func (handlers fanout) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (handlers fanout) Handle(ctx context.Context, record slog.Record) error {
	var failures []error
	for _, handler := range handlers {
		if handler.Enabled(ctx, record.Level) {
			failures = append(failures, handler.Handle(ctx, record.Clone()))
		}
	}
	return errors.Join(failures...)
}

func (handlers fanout) WithAttrs(attributes []slog.Attr) slog.Handler {
	var next fanout = make(fanout, len(handlers))
	for index, handler := range handlers {
		next[index] = handler.WithAttrs(attributes)
	}
	return next
}

func (handlers fanout) WithGroup(name string) slog.Handler {
	var next fanout = make(fanout, len(handlers))
	for index, handler := range handlers {
		next[index] = handler.WithGroup(name)
	}
	return next
}
