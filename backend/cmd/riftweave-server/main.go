// Command riftweave-server is the HTTP API (docs/backend/api.md): one POST per command, the
// session endpoints, and an OpenAPI document built from the contract. It parses, calls
// core.Run, and formats; it holds no rule and no state between calls.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/SETA1609/riftweave-core/backend/contract"
	"github.com/SETA1609/riftweave-core/backend/core"
)

// build is the server's build, set at link time (-ldflags "-X main.build=..."), dev otherwise.
var build = "dev"

// The health endpoints (docs/backend/api.md#health).
const (
	pathHealth  = "/healthz"
	pathReady   = "/readyz"
	pathVersion = "/version"
)

// The error kinds a refusal is recorded with (error.type): low cardinality, never the message,
// which may repeat what the player typed (docs/backend/observability.md#nothing-about-the-player).
const (
	kindUnknownCommand = "unknown_command"
	kindBadInput       = "bad_input"
	kindNotDefined     = "not_defined"
	kindMissingInput   = "missing_input"
	kindRule           = "rule"
	kindInternal       = "internal"
	kindNotFound       = "not_found"
	kindMethod         = "method_not_allowed"
	kindTooLarge       = "too_large"
)

// maxBody is the largest request body the server reads.
const maxBody = 1 << 20

// notServed are the commands the HTTP API never serves: dev operations are for the CLIs only
// (docs/backend/dev-mode.md).
var notServed = []string{"dev"}

// servedCommands is every core command the HTTP API serves, sorted.
func servedCommands() []string {
	var served []string
	for _, name := range core.Commands() {
		if !slices.Contains(notServed, name) {
			served = append(served, name)
		}
	}
	return served
}

func main() {
	var dataFlag, address, telemetryFile string
	flag.StringVar(&dataFlag, "data", "", "the ruleset data directory (default: RIFTWEAVE_DATA, then ruleset/data above the working directory)")
	flag.StringVar(&address, "addr", "127.0.0.1:8080", "the address to listen on")
	flag.StringVar(&telemetryFile, "telemetry", "", "write traces, metrics, and logs to this file (docs/backend/observability.md); OTEL_EXPORTER_OTLP_ENDPOINT exports them instead")
	flag.Parse()
	var ctx context.Context = context.Background()
	observer, err := telemetryFor(ctx, telemetryMode(os.Getenv, telemetryFile), telemetryFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = observer.shutdown(ctx) }()
	directory, err := rulesetDirectory(dataFlag, os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	opened, err := core.Open(directory)
	if err != nil {
		log.Fatalf("loading %s: %v", directory, err)
	}
	handler, err := newObservedHandler(opened, filepath.Join(directory, "..", "schemas", "schema.json"), observer)
	if err != nil {
		log.Fatal(err)
	}
	var httpServer *http.Server = &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	log.Printf("riftweave-server on http://%s (ruleset %s)", address, directory)
	observer.logger.InfoContext(ctx, "riftweave-server started", attributeServiceBuild, build, attributeRules, core.RulesVersion)
	log.Fatal(httpServer.ListenAndServe())
}

// rulesetDirectory finds the data as the agent CLI does (docs/backend/cli.md).
func rulesetDirectory(dataFlag string, getenv func(string) string) (string, error) {
	if dataFlag != "" {
		return dataFlag, nil
	}
	if fromEnvironment := getenv("RIFTWEAVE_DATA"); fromEnvironment != "" {
		return fromEnvironment, nil
	}
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		var candidate string = filepath.Join(directory, "ruleset", "data")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		var parent string = filepath.Dir(directory)
		if parent == directory {
			return "", errors.New("no ruleset/data found; pass --data <dir> or set RIFTWEAVE_DATA")
		}
		directory = parent
	}
}

// api is the handler: the core, and the documents built once at start-up.
type api struct {
	core     *core.Core
	openAPI  []byte
	commands []byte
	observer *telemetry
}

// newHandler is the router with no telemetry.
func newHandler(opened *core.Core, schemaPath string) (http.Handler, error) {
	return newAPI(opened, schemaPath, quietTelemetry())
}

// newObservedHandler is the router observed: a span per request that continues an incoming
// traceparent, the HTTP server metrics, and a log line for every 4xx and 5xx.
func newObservedHandler(opened *core.Core, schemaPath string, observer *telemetry) (http.Handler, error) {
	router, err := newAPI(opened, schemaPath, observer)
	if err != nil {
		return nil, err
	}
	return otelhttp.NewHandler(logged(router, observer), serviceName,
		otelhttp.WithTracerProvider(observer.tracers), otelhttp.WithMeterProvider(observer.meters), otelhttp.WithPropagators(observer.propagator),
		otelhttp.WithSpanNameFormatter(func(_ string, request *http.Request) string { return request.Method + " " + request.URL.Path })), nil
}

// refusalKey is the request context's slot for the error kind a handler refused with.
type refusalKey struct{}

// logged writes a log line for every response of 400 or more, with its error kind and status.
func logged(next http.Handler, observer *telemetry) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var kind string
		var recorder *statusRecorder = &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request.WithContext(context.WithValue(request.Context(), refusalKey{}, &kind)))
		if recorder.status < http.StatusBadRequest {
			return
		}
		var level slog.Level = slog.LevelWarn
		if recorder.status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		observer.logger.Log(request.Context(), level, "request refused", attributeErrorType, kind, attributeStatus, recorder.status)
	})
}

// statusRecorder keeps the status a handler wrote.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

// refused records the error kind for the request's log line and its span.
func refused(request *http.Request, kind string) {
	if slot, ok := request.Context().Value(refusalKey{}).(*string); ok {
		*slot = kind
	}
}

// newAPI builds the OpenAPI document and the command list, and returns the router.
func newAPI(opened *core.Core, schemaPath string, observer *telemetry) (http.Handler, error) {
	document, err := buildOpenAPI(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("OpenAPI document: %w", err)
	}
	type entry struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	}
	var entries []entry = []entry{}
	for _, name := range servedCommands() {
		var file struct {
			Summary string `json:"summary"`
		}
		content, err := fs.ReadFile(contract.Files, name+".json")
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(content, &file); err != nil {
			return nil, err
		}
		entries = append(entries, entry{Name: name, Summary: file.Summary})
	}
	list, err := json.Marshal(entries)
	if err != nil {
		return nil, err
	}
	return &api{core: opened, openAPI: document, commands: list, observer: observer}, nil
}

// sessionPaths are the session endpoints and the operation each fills in.
var sessionPaths = map[string]string{"/v1/sessions": "start", "/v1/sessions/step": "step", "/v1/sessions/replay": "replay"}

func (server *api) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var path string = request.URL.Path
	switch {
	case path == pathHealth:
		if allowed(writer, request, http.MethodGet) {
			write(writer, http.StatusOK, []byte(`{"status":"ok"}`))
		}
	case path == pathReady:
		if allowed(writer, request, http.MethodGet) {
			write(writer, http.StatusOK, []byte(`{"status":"ready"}`)) // the ruleset loaded before the server listened
		}
	case path == pathVersion:
		if allowed(writer, request, http.MethodGet) {
			body, _ := json.Marshal(map[string]string{"build": build, "rules_version": core.RulesVersion})
			write(writer, http.StatusOK, body)
		}
	case path == "/openapi.json":
		if allowed(writer, request, http.MethodGet) {
			write(writer, http.StatusOK, server.openAPI)
		}
	case path == "/v1/commands":
		if allowed(writer, request, http.MethodGet) {
			write(writer, http.StatusOK, server.commands)
		}
	case strings.HasPrefix(path, "/v1/commands/"):
		if !allowed(writer, request, http.MethodPost) {
			return
		}
		var name string = strings.TrimPrefix(path, "/v1/commands/")
		if slices.Contains(notServed, name) {
			refused(request, kindNotFound)
			fail(writer, http.StatusNotFound, fmt.Sprintf("%s is not served over HTTP; use the CLI with --dev", name))
			return
		}
		if body, ok := readBody(writer, request); ok {
			server.run(writer, request, name, body)
		}
	case sessionPaths[path] != "":
		if !allowed(writer, request, http.MethodPost) {
			return
		}
		body, ok := readBody(writer, request)
		if !ok {
			return
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
			refused(request, kindBadInput)
			fail(writer, http.StatusBadRequest, "the body is not a JSON object")
			return
		}
		operation, _ := json.Marshal(sessionPaths[path])
		fields["operation"] = operation
		input, _ := json.Marshal(fields)
		server.run(writer, request, "session", input)
	default:
		refused(request, kindNotFound)
		fail(writer, http.StatusNotFound, fmt.Sprintf("no endpoint %s", path))
	}
}

// run calls the core inside a span named for the command, measures it, and maps its errors to a
// status (docs/backend/api.md#status-codes). A refusal's span and measurement carry its kind only.
func (server *api) run(writer http.ResponseWriter, request *http.Request, name string, body []byte) {
	var command attribute.KeyValue = attribute.String(attributeCommand, name)
	ctx, span := server.observer.tracer.Start(request.Context(), spanCommandPrefix+name, trace.WithAttributes(command))
	defer span.End()
	var started time.Time = time.Now()
	output, err := server.core.Run(name, body)
	var kind string = refusalKind(err)
	var measured []attribute.KeyValue = []attribute.KeyValue{command}
	if kind != "" {
		measured = append(measured, attribute.String(attributeErrorType, kind))
		span.SetAttributes(attribute.String(attributeErrorType, kind))
		span.SetStatus(codes.Error, kind)
		refused(request, kind)
	}
	server.observer.duration.Record(ctx, time.Since(started).Seconds(), metric.WithAttributes(measured...))
	switch kind {
	case "":
		write(writer, http.StatusOK, output)
	case kindUnknownCommand:
		fail(writer, http.StatusNotFound, err.Error())
	case kindBadInput:
		fail(writer, http.StatusBadRequest, err.Error())
	case kindNotDefined:
		write(writer, http.StatusNotFound, output)
	case kindMissingInput:
		write(writer, http.StatusBadRequest, output)
	case kindRule:
		write(writer, http.StatusUnprocessableEntity, output)
	default:
		fail(writer, http.StatusInternalServerError, err.Error())
	}
}

// refusalKind names the kind of a core error, "" for none; the most specific wins.
func refusalKind(err error) string {
	var notDefined *core.NotDefinedError
	var missing *core.MissingInputError
	if err == nil {
		return ""
	}
	if errors.Is(err, core.ErrUnknownCommand) {
		return kindUnknownCommand
	}
	if errors.Is(err, core.ErrInput) {
		return kindBadInput
	}
	if errors.As(err, &notDefined) {
		return kindNotDefined
	}
	if errors.As(err, &missing) {
		return kindMissingInput
	}
	if errors.Is(err, core.ErrRule) {
		return kindRule
	}
	return kindInternal
}

func allowed(writer http.ResponseWriter, request *http.Request, method string) bool {
	if request.Method == method {
		return true
	}
	writer.Header().Set("Allow", method)
	refused(request, kindMethod)
	fail(writer, http.StatusMethodNotAllowed, fmt.Sprintf("%s takes %s", request.URL.Path, method))
	return false
}

func readBody(writer http.ResponseWriter, request *http.Request) ([]byte, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, maxBody))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		refused(request, kindTooLarge)
		fail(writer, http.StatusRequestEntityTooLarge, "the body is over 1 MiB")
		return nil, false
	}
	if err != nil {
		refused(request, kindBadInput)
		fail(writer, http.StatusBadRequest, err.Error())
		return nil, false
	}
	return body, true
}

func write(writer http.ResponseWriter, status int, body []byte) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_, _ = writer.Write(body)
}

func fail(writer http.ResponseWriter, status int, message string) {
	body, _ := json.Marshal(map[string]string{"error": message})
	write(writer, status, body)
}
