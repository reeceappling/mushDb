package gotel

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupSDK(ctx context.Context) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider()
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider()
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider()
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider() (*trace.TracerProvider, error) {
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return tracerProvider, nil
}

func newMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {
	logExporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}

//	"github.com/caarlos0/env" // TODO: validate ok to use this package
//	"go.opentelemetry.io/contrib/bridges/otelzap"
//	"go.opentelemetry.io/otel"
//	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
//	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
//	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
//	otelmetric "go.opentelemetry.io/otel/metric"
//	"go.opentelemetry.io/otel/sdk/log"
//	"go.opentelemetry.io/otel/sdk/metric"
//	"go.opentelemetry.io/otel/sdk/resource"
//	"go.opentelemetry.io/otel/sdk/trace"
//	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
//	oteltrace "go.opentelemetry.io/otel/trace"
//	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
//	"os"
//)
//
//// TODO: SEE THIS! https://www.lucavallin.com/blog/opentelemetry-a-guide-to-observability-with-go
//
//// Config holds the configuration for the telemetry.
//type Config struct {
//	ServiceName    string `env:"SERVICE_NAME"      envDefault:"gotel"`
//	ServiceVersion string `env:"SERVICE_VERSION"   envDefault:"0.0.1"`
//	Enabled        bool   `env:"TELEMETRY_ENABLED" envDefault:"true"`
//}
//
//// NewConfigFromEnv creates a new telemetry config from the environment.
//func NewConfigFromEnv() (Config, error) {
//	telem := Config{}
//	if err := env.Parse(&telem); err != nil {
//		return Config{}, fmt.Errorf("failed to parse telemetry config: %w", err)
//	}
//
//	return telem, nil
//}
//
//type Metric struct {
//	Name        string
//	Description string
//	Unit        string
//}
//
//// TelemetryProvider is an interface for the telemetry provider.
//type TelemetryProvider interface {
//	GetServiceName() string
//	LogInfo(args ...interface{})
//	LogErrorln(args ...interface{})
//	LogFatalln(args ...interface{})
//	MeterInt64Histogram(metric Metric) (otelmetric.Int64Histogram, error)
//	MeterInt64UpDownCounter(metric Metric) (otelmetric.Int64UpDownCounter, error)
//	TraceStart(ctx context.Context, name string) (context.Context, oteltrace.Span)
//	LogRequest() gin.HandlerFunc            // TODO: fix!
//	MeterRequestDuration() gin.HandlerFunc  // TODO: fix!
//	MeterRequestsInFlight() gin.HandlerFunc // TODO: fix!
//	Shutdown(ctx context.Context)
//}
//
//// Telemetry is a wrapper around the OpenTelemetry logger, meter, and tracer.
//type Telemetry struct {
//	lp     *log.LoggerProvider
//	mp     *metric.MeterProvider
//	tp     *trace.TracerProvider
//	log    *zap.SugaredLogger
//	meter  otelmetric.Meter
//	tracer oteltrace.Tracer
//	cfg    Config
//}
//
//// NewTelemetry creates a new telemetry instance.
//func NewTelemetry(ctx context.Context, cfg Config) (*Telemetry, error) {
//	rp := newResource(cfg.ServiceName, cfg.ServiceVersion)
//
//	lp, err := newLoggerProvider(ctx, rp)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create logger: %w", err)
//	}
//
//	logger := zap.New(
//		zapcore.NewTee(
//			zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
//			otelzap.NewCore(cfg.ServiceName, otelzap.WithLoggerProvider(lp)),
//		),
//	)
//
//	mp, err := newMeterProvider(ctx, rp)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create meter: %w", err)
//	}
//	meter := mp.Meter(cfg.ServiceName)
//
//	tp, err := newTracerProvider(ctx, rp)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create tracer: %w", err)
//	}
//	tracer := tp.Tracer(cfg.ServiceName)
//
//	return &Telemetry{
//		lp:     lp,
//		mp:     mp,
//		tp:     tp,
//		log:    logger.Sugar(),
//		meter:  meter,
//		tracer: tracer,
//		cfg:    cfg,
//	}, nil
//}
//
//// GetServiceName returns the name of the service.
//func (t *Telemetry) GetServiceName() string {
//	return t.cfg.ServiceName
//}
//
//// LogInfo logs a message at the info level.
//func (t *Telemetry) LogInfo(args ...interface{}) {
//	t.log.Info(args...)
//}
//
//// LogErrorln logs a message and then calls os.Exit(1).
//func (t *Telemetry) LogErrorln(args ...interface{}) {
//	t.log.Errorln(args...)
//}
//
//// LogFatalln logs a message and then calls os.Exit(1).
//func (t *Telemetry) LogFatalln(args ...interface{}) {
//	t.log.Fatalln(args...)
//}
//
//// MeterInt64Histogram creates a new int64 histogram metric.
//func (t *Telemetry) MeterInt64Histogram(metric Metric) (otelmetric.Int64Histogram, error) { //nolint:ireturn
//	histogram, err := t.meter.Int64Histogram(
//		metric.Name,
//		otelmetric.WithDescription(metric.Description),
//		otelmetric.WithUnit(metric.Unit),
//	)
//
//	if err != nil {
//		return nil, fmt.Errorf("failed to create histogram: %w", err)
//	}
//
//	return histogram, nil
//}
//
//// MeterInt64UpDownCounter creates a new int64 up down counter metric.
//func (t *Telemetry) MeterInt64UpDownCounter(metric Metric) (otelmetric.Int64UpDownCounter, error) { //nolint:ireturn
//	counter, err := t.meter.Int64UpDownCounter(
//		metric.Name,
//		otelmetric.WithDescription(metric.Description),
//		otelmetric.WithUnit(metric.Unit),
//	)
//
//	if err != nil {
//		return nil, fmt.Errorf("failed to create counter: %w", err)
//	}
//
//	return counter, nil
//}
//
//// TraceStart starts a new span with the given name. The span must be ended by calling End.
//func (t *Telemetry) TraceStart(ctx context.Context, name string) (context.Context, oteltrace.Span) { //nolint:ireturn
//	//nolint: spancheck
//	return t.tracer.Start(ctx, name)
//}
//
//// Shutdown shuts down the logger, meter, and tracer.
//func (t *Telemetry) Shutdown(ctx context.Context) {
//	t.lp.Shutdown(ctx)
//	t.mp.Shutdown(ctx)
//	t.tp.Shutdown(ctx)
//}
//
//// newLoggerProvider creates a new logger provider with the OTLP gRPC exporter.
//func newLoggerProvider(ctx context.Context, res *resource.Resource) (*log.LoggerProvider, error) {
//	exporter, err := otlploggrpc.New(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
//	}
//
//	processor := log.NewBatchProcessor(exporter)
//	lp := log.NewLoggerProvider(
//		log.WithProcessor(processor),
//		log.WithResource(res),
//	)
//
//	return lp, nil
//}
//
//// newMeterProvider creates a new meter provider with the OTLP gRPC exporter.
//func newMeterProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
//	exporter, err := otlpmetricgrpc.New(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
//	}
//
//	mp := metric.NewMeterProvider(
//		metric.WithReader(metric.NewPeriodicReader(exporter)),
//		metric.WithResource(res),
//	)
//	otel.SetMeterProvider(mp)
//
//	return mp, nil
//}
//
//// newTracerProvider creates a new tracer provider with the OTLP gRPC exporter.
//func newTracerProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
//	exporter, err := otlptracegrpc.New(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
//	}
//
//	// Create Resource
//	tp := trace.NewTracerProvider(
//		trace.WithBatcher(exporter),
//		trace.WithResource(res),
//	)
//	otel.SetTracerProvider(tp)
//
//	return tp, nil
//}
//
//// newResource creates a new OTEL resource with the service name and version.
//func newResource(serviceName string, serviceVersion string) *resource.Resource {
//	hostName, _ := os.Hostname()
//
//	return resource.NewWithAttributes(
//		semconv.SchemaURL,
//		semconv.ServiceName(serviceName),
//		semconv.ServiceVersion(serviceVersion),
//		semconv.HostName(hostName),
//	)
//}
