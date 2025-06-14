package signozutils

import (
	"context"
	"log"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

type Tracer interface {
	Cleanup(ctx context.Context)
	StartSpan(ctx context.Context, spanName string) (context.Context, oteltrace.Span)
}

type tracer struct {
	collectorURL   string
	serviceName    string
	insecure       string
	tracerShutdown func(context.Context) error
	meterProvider  *metricsdk.MeterProvider
	meter          api.Meter
	otelTracer     oteltrace.Tracer
}

func InitTracer(collectorURL string, serviceName string, insecure string) Tracer {
	tracer := &tracer{
		collectorURL: collectorURL,
		serviceName:  serviceName,
		insecure:     insecure,
	}

	otelTracer := otel.Tracer(tracer.serviceName)
	tracer.otelTracer = otelTracer

	tracerShutdown := tracer.initTracer()
	tracer.tracerShutdown = tracerShutdown

	meterProvider := tracer.initMeter()
	tracer.meterProvider = meterProvider

	meter := meterProvider.Meter(tracer.serviceName)
	tracer.meter = meter

	tracer.generateMetrics()

	return tracer
}

func (t *tracer) StartSpan(ctx context.Context, spanName string) (context.Context, oteltrace.Span) {
	tracerCtx, span := t.otelTracer.Start(ctx, spanName)
	return tracerCtx, span
}

func (t *tracer) Cleanup(ctx context.Context) {
	if t.tracerShutdown != nil {
		if err := t.tracerShutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}
	if t.meterProvider != nil {
		if err := t.meterProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}
	log.Println("Tracer cleaned up")
}

func (t *tracer) initTracer() func(context.Context) error {

	var secureOption otlptracegrpc.Option

	if strings.ToLower(t.insecure) == "false" || t.insecure == "0" || strings.ToLower(t.insecure) == "f" {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(t.collectorURL),
		),
	)

	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(t.serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Fatalf("Could not set resources: %v", err)
	}

	// TODO: For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	log.Println("Tracer initialized")

	return exporter.Shutdown
}

func (t *tracer) initMeter() *metricsdk.MeterProvider {
	secureOption := otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(t.insecure) > 0 {
		secureOption = otlpmetricgrpc.WithInsecure()
	}

	exporter, err := otlpmetricgrpc.New(
		context.Background(),
		secureOption,
		otlpmetricgrpc.WithEndpoint(t.collectorURL),
	)

	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", t.serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Fatalf("Could not set resources: %v", err)
	}

	// Register the exporter with an SDK via a periodic reader.
	provider := metricsdk.NewMeterProvider(
		metricsdk.WithResource(res),
		metricsdk.WithReader(metricsdk.NewPeriodicReader(exporter)),
	)

	return provider
}

func (t *tracer) generateMetrics() {
	go exceptionsCounter(t.meter)
	go pageFaultsCounter(t.meter)
	go requestDurationHistogram(t.meter)
	go roomTemperatureGauge(t.meter)
	go itemsInQueueUpDownCounter(t.meter)
	go processHeapSizeUpDownCounter(t.meter)
}
