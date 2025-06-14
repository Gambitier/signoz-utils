# SigNoz Utils

A Go utility library for working with SigNoz observability platform, providing easy-to-use OpenTelemetry instrumentation helpers.

## Overview

SigNoz Utils is a Go package that simplifies the integration of OpenTelemetry metrics and tracing into your applications. It provides a set of utility functions and helpers for common observability patterns.

## Features

- Easy-to-use OpenTelemetry instrumentation
- Support for various metric types
- GRPC integration support
- Secure and insecure collector connections

## Installation

```bash
go get github.com/Gambitier/signoz-utils
```

## Usage

### Tracing

Here's how to use the tracing functionality:

```go
package main

import (
    "os"
    "context"

    signozutils "github.com/Gambitier/signoz-utils"
)

var (
	collectorURL = os.Getenv("COLLECTOR_URL")
	serviceName  = os.Getenv("SERVICE_NAME")
	insecureMode = os.Getenv("INSECURE")
)

func init() {
	if insecureMode == "" {
		insecureMode = "true"
	}

	if collectorURL == "" {
		collectorURL = "localhost:4317" // NOTE: DO NOT USE http or https before
	}

	if serviceName == "" {
		serviceName = "my-service"
	}
}

func main() {
    // Initialize the tracer
    tracer := signozutils.InitTracer(
        collectorURL,  // collector URL
        serviceName,   // service name
        insecureMode,  // insecure flag (true/false)
    )
    defer tracer.Cleanup(context.Background())
.
.
}
```

Create and start tracing span

```go
    // Create a new span
    ctx, span := tracer.StartSpan(context.Background(), "operation-name")
    defer span.End()

    // Use the context in your operations
    // The context will contain the trace information
```

### Metrics

Here's how to use the metrics functionality:

```go
package main

import (
    "github.com/gorilla/mux"
	otelMiddleware "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)
.
.
.
	router := mux.NewRouter()
	router.Use(otelMiddleware.Middleware(serviceName))
.
.
```

## Dependencies

- Go 1.23.3 or higher
- OpenTelemetry SDK v1.36.0
- Google gRPC v1.73.0

## References

- [Go OpenTelemetry Instrumentation](https://signoz.io/docs/instrumentation/opentelemetry-golang/)
- [Bookstore REST API using Gin and Gorm](https://github.com/SigNoz/sample-golang-app)
- [distributed-tracing-golang-sample](https://github.com/SigNoz/distributed-tracing-golang-sample/)

## License

This project is licensed under the MIT License - see the LICENSE file for details.
