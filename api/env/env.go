package env

import "context"

type envKey string

const envCtxKey envKey = "environment"
const Prod = "prod"
const Cert = "cert" // TODO: qual, cert?
const Dev = "dev"   // TODO: qual, cert?

func GetEnv(ctx context.Context) string {
	if out, exists := ctx.Value(envCtxKey).(string); exists {
		return out
	}
	return Prod
}
func SetEnv(ctx context.Context, env string) context.Context {
	return context.WithValue(ctx, envCtxKey, env)
}
func IfDev(ctx context.Context, doIfDev func() error) error { // TODO: use this everywhere needed
	if GetEnv(ctx) == Dev {
		return doIfDev()
	}
	return nil
}
func IfCert(ctx context.Context, doIfProd func() error) error { // TODO: use this everywhere needed
	if GetEnv(ctx) == Cert {
		return doIfProd()
	}
	return nil
}
func IfProd(ctx context.Context, doIfProd func() error) error { // TODO: use this everywhere needed
	if GetEnv(ctx) == Prod {
		return doIfProd()
	}
	return nil
}
func IfNotProd(ctx context.Context, doIfProd func() error) error { // TODO: use this everywhere needed
	if GetEnv(ctx) != Prod {
		return doIfProd()
	}
	return nil
}
func LogIfDev(ctx context.Context, toLog string) { // TODO: use this everywhere needed
	if GetEnv(ctx) == Dev {
		println(toLog)
	}
}
func LogAlways(toLog string) { // TODO: use this everywhere needed
	println(toLog)
}
