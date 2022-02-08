package otel

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// OutputType are supported otel outputs.
type OutputType int

// Supported otel outputs.
const (
	// IO is a IO output type like logger or File.
	IO OutputType = iota

	// GRPC is a protocol we supported to send to supported GRPC endpoints
	GRPC
)

// Config holds the default required values to open a set OTEL pipeline
//
// Writer just used for IO output in this case APIKey and URL can be empty
// APIKey and URL are using fo GRPC output in this case Writer can be nil
type Config struct {
	ServiceName       string
	ServiceVersion    string
	ServiceInstanceID string
	Writer            io.Writer
	APIKey            string
	URL               string
}

func (c *Config) resource(ctx context.Context) (*resource.Resource, error) {
	defaultResource, _ := resource.New(ctx)
	resource, err := resource.Merge(
		defaultResource,
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.ServiceName),
			semconv.ServiceVersionKey.String(c.ServiceVersion),
			semconv.ServiceInstanceIDKey.String(c.ServiceInstanceID),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("could not create resource: %w", err)
	}

	return resource, nil
}

// Exporter exposes a common interface to perform
// otel export pipeline to different supported outputs
type Exporter interface {
	ExportPipeline(context.Context) (*trace.TracerProvider, error)
}

type ioOutput struct {
	*Config
}

// Export implements the Exporter interface for IO output.
func (c *ioOutput) ExportPipeline(ctx context.Context) (*trace.TracerProvider, error) {
	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(c.Config.Writer),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create exporter: %w", err)
	}

	resource, _ := c.Config.resource(ctx)
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		//trace.
		trace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}

type grpcOutput struct {
	*Config
}

// Export implements the Exporter interface for GRPC output.
func (g *grpcOutput) ExportPipeline(ctx context.Context) (*trace.TracerProvider, error) {
	var headers = map[string]string{
		"api-key": g.Config.APIKey,
	}

	creds := credentials.NewClientTLSFromCert(nil, "")

	var clientOpts = []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(g.Config.URL),
		otlptracegrpc.WithTLSCredentials(creds),
		otlptracegrpc.WithReconnectionPeriod(2 * time.Second),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
		otlptracegrpc.WithTimeout(30 * time.Second),
		otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithCompressor("gzip"),
	}

	otlpExporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(clientOpts...))
	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	resource, _ := g.Config.resource(ctx)
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(otlpExporter,
			trace.WithBatchTimeout(5*time.Second),
			trace.WithExportTimeout(5*time.Second),
			trace.WithMaxQueueSize(10000),
			trace.WithMaxExportBatchSize(100000),
		),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(resource),
	)
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}

// NewExporter builds the otel exporter pipeline as specified.
func NewExporter(outputType OutputType, c *Config) Exporter {
	switch outputType {
	case IO:
		return &ioOutput{
			Config: c,
		}
	case GRPC:
		return &grpcOutput{
			Config: c,
		}
	}

	return nil
}

// NewENVConfig constructs a configuration object from
// the values found on the environment.
func NewENVConfig() *Config {
	return &Config{
		ServiceName:       os.Getenv("OTEL_SERVICE_NAME"),
		ServiceVersion:    os.Getenv("OTEL_SERVICE_VERSION"),
		ServiceInstanceID: os.Getenv("OTEL_SERVICE_ID"),
		Writer:            nil,
		APIKey:            os.Getenv("OTEL_GRPC_API_KEY"),
		URL:               os.Getenv("OTEL_GRPC_URL"),
	}
}
