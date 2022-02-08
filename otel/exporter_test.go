package otel

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExporter_TraceSpan(t *testing.T) {
	setEnv()
	defer unsetEnv()

	var ioWriter bytes.Buffer
	c := NewENVConfig()
	c.Writer = &ioWriter

	exporter := NewExporter(IO, c)
	pipeline, err := exporter.ExportPipeline(context.TODO())
	defer pipeline.Shutdown(context.TODO())

	assert.Nil(t, err)
	assert.Empty(t, ioWriter)

	_, span := pipeline.Tracer("sample").Start(context.TODO(), "sample span")
	assert.True(t, span.IsRecording())
	span.End()
	assert.False(t, span.IsRecording())
	err = pipeline.ForceFlush(context.TODO())

	assert.Nil(t, err)
	assert.NotEmpty(t, ioWriter)
	assert.NotEmpty(t, pipeline)

}

func TestExporter_CreateOTELResources(t *testing.T) {
	setEnv()
	defer unsetEnv()

	c := NewENVConfig()
	c.Writer = os.Stdout
	resource, _ := c.resource(context.TODO())
	attr := resource.Attributes()

	assert.Equal(t, "sampleServiceID", attr[0].Value.AsString())
	assert.Equal(t, "sampleServiceName", attr[1].Value.AsString())
	assert.Equal(t, "v1.0.0.0", attr[2].Value.AsString())

}

func TestExporter_GetPipelineWithOutputTypeIO(t *testing.T) {
	setEnv()
	defer unsetEnv()

	c := NewENVConfig()
	c.Writer = os.Stdout
	exporter := NewExporter(IO, c)
	pipeline, err := exporter.ExportPipeline(context.TODO())
	defer pipeline.Shutdown(context.TODO())

	assert.Nil(t, err)
	assert.NotEmpty(t, pipeline)

}

func TestExporter_GetPipelineWithOutputTypeGRPC(t *testing.T) {
	setEnv()
	defer unsetEnv()

	c := NewENVConfig()
	exporter := NewExporter(GRPC, c)
	pipeline, err := exporter.ExportPipeline(context.TODO())
	defer pipeline.Shutdown(context.TODO())

	assert.Nil(t, err)
	assert.NotEmpty(t, pipeline)

}

func setEnv() {
	os.Setenv("OTEL_SERVICE_NAME", "sampleServiceName")
	os.Setenv("OTEL_SERVICE_VERSION", "v1.0.0.0")
	os.Setenv("OTEL_SERVICE_ID", "sampleServiceID")
	os.Setenv("OTEL_GRPC_API_KEY", "sampleApiKey")
	os.Setenv("OTEL_GRPC_URL", "otlp.nr-data.net:4317")
}

func unsetEnv() {
	os.Unsetenv("OTEL_SERVICE_NAME")
	os.Unsetenv("OTEL_SERVICE_VERSION")
	os.Unsetenv("OTEL_SERVICE_ID")
	os.Unsetenv("OTEL_GRPC_API_KEY")
	os.Unsetenv("OTEL_GRPC_URL")
}
