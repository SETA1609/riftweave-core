package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/SETA1609/riftweave-core/backend/core"
)

// The server is observable (BE_S12_02, docs/backend/observability.md): a request span with a
// child span per core.Run, the command duration metric, and log records carrying the trace id,
// all checked here through in-memory exporters.

// memoryLogs keeps every log record the server emits.
type memoryLogs struct {
	mutex   sync.Mutex
	records []sdklog.Record
}

func (logs *memoryLogs) Export(_ context.Context, records []sdklog.Record) error {
	logs.mutex.Lock()
	defer logs.mutex.Unlock()
	for _, record := range records {
		logs.records = append(logs.records, record.Clone())
	}
	return nil
}
func (logs *memoryLogs) Shutdown(context.Context) error   { return nil }
func (logs *memoryLogs) ForceFlush(context.Context) error { return nil }

// observedServer is the server with telemetry sent to memory.
type observedServer struct {
	served *httptest.Server
	spans  *tracetest.InMemoryExporter
	reader *sdkmetric.ManualReader
	logs   *memoryLogs
	lines  *bytes.Buffer
}

func observed(t *testing.T) observedServer {
	t.Helper()
	opened, err := core.Open(filepath.Join(repo, "ruleset", "data"))
	if err != nil {
		t.Fatal(err)
	}
	var server observedServer = observedServer{spans: tracetest.NewInMemoryExporter(), reader: sdkmetric.NewManualReader(), logs: &memoryLogs{}, lines: &bytes.Buffer{}}
	telemetry, err := newTelemetry(context.Background(), telemetryOptions{
		spans: server.spans, metrics: server.reader, logs: server.logs, lines: server.lines, build: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = telemetry.shutdown(context.Background()) })
	handler, err := newObservedHandler(opened, filepath.Join(repo, "ruleset", "schemas", "schema.json"), telemetry)
	if err != nil {
		t.Fatal(err)
	}
	server.served = httptest.NewServer(handler)
	t.Cleanup(server.served.Close)
	return server
}

// incoming is a caller's W3C traceparent: its trace id and its span.
const (
	incomingTrace = "4bf92f3577b34da6a3ce929d0e0e4736"
	incoming      = "00-" + incomingTrace + "-00f067aa0ba902b7-01"
)

func postTraced(t *testing.T, server observedServer, path, body string) int {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, server.served.URL+path, strings.NewReader(body))
	request.Header.Set("traceparent", incoming)
	response, err := server.served.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	return response.StatusCode
}

// A request is a server span that continues the caller's trace, with one child span for the
// core command it ran.
func TestARequestIsASpanWithItsCommandAsAChild(t *testing.T) {
	server := observed(t)
	if status := postTraced(t, server, "/v1/commands/resolution", `{"actor":{"skills":{"lockpick":40},"attributes":{"lck":5}},"skill":"lockpick","roll":23}`); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	var request, command *tracetest.SpanStub
	for index, span := range server.spans.GetSpans() {
		switch {
		case span.SpanKind == trace.SpanKindServer:
			request = &server.spans.GetSpans()[index]
		case span.Name == "riftweave.command resolution":
			command = &server.spans.GetSpans()[index]
		}
	}
	if request == nil || command == nil {
		t.Fatalf("want a server span and a command span, got %+v", server.spans.GetSpans())
	}
	if request.SpanContext.TraceID().String() != incomingTrace {
		t.Errorf("the request span's trace %s, want the caller's %s", request.SpanContext.TraceID(), incomingTrace)
	}
	if command.Parent.SpanID() != request.SpanContext.SpanID() || command.SpanContext.TraceID() != request.SpanContext.TraceID() {
		t.Errorf("the command span is not the request span's child: %+v", command.Parent)
	}
	var named bool
	for _, attribute := range command.Attributes {
		named = named || (string(attribute.Key) == "riftweave.command.name" && attribute.Value.AsString() == "resolution")
	}
	if !named {
		t.Errorf("the command span has no riftweave.command.name: %v", command.Attributes)
	}
}

// Each command call is one measurement of riftweave.command.duration, by command.
func TestEachCommandIsMeasured(t *testing.T) {
	server := observed(t)
	postTraced(t, server, "/v1/commands/resolution", `{"actor":{"skills":{"lockpick":40},"attributes":{"lck":5}},"skill":"lockpick","roll":23}`)
	var collected metricdata.ResourceMetrics
	if err := server.reader.Collect(context.Background(), &collected); err != nil {
		t.Fatal(err)
	}
	for _, scope := range collected.ScopeMetrics {
		for _, measured := range scope.Metrics {
			if measured.Name != "riftweave.command.duration" {
				continue
			}
			histogram, ok := measured.Data.(metricdata.Histogram[float64])
			if !ok || len(histogram.DataPoints) != 1 || histogram.DataPoints[0].Count != 1 {
				t.Fatalf("riftweave.command.duration: %+v", measured.Data)
			}
			if name, _ := histogram.DataPoints[0].Attributes.Value("riftweave.command.name"); name.AsString() != "resolution" {
				t.Errorf("measured for %q, want resolution", name.AsString())
			}
			return
		}
	}
	t.Fatalf("no riftweave.command.duration in %+v", collected)
}

// A refused call is logged with its error kind and the request's trace id, both as an
// OpenTelemetry record and as a JSON line, and never with what the player typed.
func TestARefusalIsLoggedWithTheTraceAndNoPlayerText(t *testing.T) {
	server := observed(t)
	const typed = "zz-the-players-own-words"
	if status := postTraced(t, server, "/v1/commands/resolution", `{"actor":{"skills":{"`+typed+`":40},"attributes":{"lck":5}},"skill":"lockpick","roll":23}`); status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d, want 422", status)
	}
	server.logs.mutex.Lock()
	defer server.logs.mutex.Unlock()
	var traced bool
	for _, record := range server.logs.records {
		traced = traced || record.TraceID().String() == incomingTrace
		if strings.Contains(record.Body().String(), typed) {
			t.Errorf("a log record carries the player's text: %s", record.Body().String())
		}
	}
	if !traced {
		t.Errorf("no log record carries the trace %s: %d records", incomingTrace, len(server.logs.records))
	}
	if lines := server.lines.String(); !strings.Contains(lines, `"trace_id":"`+incomingTrace+`"`) || strings.Contains(lines, typed) {
		t.Errorf("the JSON log lines: %s", lines)
	}
	for _, span := range server.spans.GetSpans() {
		for _, attribute := range span.Attributes {
			if strings.Contains(attribute.Value.Emit(), typed) {
				t.Errorf("span %s carries the player's text in %s", span.Name, attribute.Key)
			}
		}
	}
}

// The health endpoints answer, and the OpenAPI document lists them.
func TestTheHealthEndpointsAnswerAndAreDocumented(t *testing.T) {
	served := server(t)
	for path, want := range map[string]map[string]any{
		"/healthz": {"status": "ok"},
		"/readyz":  {"status": "ready"},
		"/version": {"build": "dev", "rules_version": core.RulesVersion},
	} {
		status, body := call(t, served, http.MethodGet, path, nil)
		for field, value := range want {
			if status != http.StatusOK || body[field] != value {
				t.Errorf("GET %s: %d %v, want %s %v", path, status, body, field, value)
			}
		}
	}
	_, document := call(t, served, http.MethodGet, "/openapi.json", nil)
	paths, _ := document["paths"].(map[string]any)
	for _, path := range []string{"/healthz", "/readyz", "/version"} {
		if _, listed := paths[path]; !listed {
			t.Errorf("the OpenAPI document has no %s", path)
		}
	}
}

// With no OTEL_EXPORTER_OTLP_ENDPOINT and no --telemetry, the server exports nothing.
func TestTelemetryIsOffUnlessAskedFor(t *testing.T) {
	environment := map[string]string{}
	getenv := func(key string) string { return environment[key] }
	if mode := telemetryMode(getenv, ""); mode != telemetryOff {
		t.Errorf("no endpoint and no file: %v, want off", mode)
	}
	if mode := telemetryMode(getenv, "diagnostics.json"); mode != telemetryFile {
		t.Errorf("--telemetry diagnostics.json: %v, want file", mode)
	}
	environment["OTEL_EXPORTER_OTLP_ENDPOINT"] = "http://localhost:4318"
	if mode := telemetryMode(getenv, ""); mode != telemetryExport {
		t.Errorf("an endpoint: %v, want export", mode)
	}
	environment["OTEL_SDK_DISABLED"] = "true"
	if mode := telemetryMode(getenv, "diagnostics.json"); mode != telemetryOff {
		t.Errorf("OTEL_SDK_DISABLED: %v, want off", mode)
	}
}
