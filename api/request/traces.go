package request

//func initTracer(ctx context.Context, serviceName string) {
//	// Inside initTracer:
//	exporter, _ := otlptracehttp.New(ctx)
//	res, _ := resource.New(ctx, resource.WithAttributes(semconv.ServiceNameKey.String(serviceName)))
//	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))
//	otel.SetTracerProvider(tp)
//	// ... use defer tp.Shutdown(ctx) in main
//}

//const TraceIdHeader = "X-Trace-Id" // TODO: ok?
//const traceIdCtxKey ctxKey = "traceIdCtxKey"
//
//func WithTraceId(ctx context.Context, incomingHeaders http.Header, w http.ResponseWriter) context.Context { // TODO: use?!
//	idToPass := incomingHeaders.Get(TraceIdHeader)
//	if idToPass == "" {
//		// create new trace id
//		idToPass = uuid.New().String()
//	}
//	w.Header().Set(TraceIdHeader, idToPass)
//	return context.WithValue(ctx, traceIdCtxKey, idToPass)
//}
//
//func GetTraceId(ctx context.Context) (string, error) { // TODO: use?!
//	out, ok := ctx.Value(traceIdCtxKey).(string)
//	if !ok {
//		return "", errors.New("no trace id found")
//	}
//	return out, nil
//}
