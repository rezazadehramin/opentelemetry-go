// Package otel handles common and reusable logic
// to interact with New Relic and other gRPC supported endpoint
// from go applications based on Open Telemetry protocol.
//
// To authenticate with new relic make sure to populate
// the following env variables.
//
// - OTEL_GRPC_API_KEY=
// - OTEL_GRPC_URL=otlp.nr-data.net:4317
//
// otel needs some other config to read better in visualization applications like NewRelic
// for this you should populate these envs too
// - OTEL_SERVICE_NAME
// - OTEL_SERVICE_VERSION
// - OTEL_SERVICE_ID
// you can export otel output in to your console output, for this purpose
// you need to set output type toIO
// The application returned already contains a configured
package otel
