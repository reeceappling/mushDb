package env

import "context"

type envKey string

const envCtxKey envKey = "environment"
const Prod = "prod"
const Dev = "dev" // TODO: qual, cert?

func GetEnv(ctx context.Context) string {
	if out, exists := ctx.Value(envCtxKey).(string); exists {
		return out
	}
	return Prod
}
func SetEnv(ctx context.Context, env string) context.Context {
	return context.WithValue(ctx, envCtxKey, env)
}
func IfDev(ctx context.Context, doIfDev func()) { // TODO: use this everywhere needed
	if GetEnv(ctx) == Dev {
		doIfDev()
	}
}
func LogIfDev(ctx context.Context, toLog string) { // TODO: use this everywhere needed
	if GetEnv(ctx) == Dev {
		println(toLog)
	}
}
func LogAlways(toLog string) { // TODO: use this everywhere needed
	println(toLog)
}
func IfProd(ctx context.Context, doIfProd func()) { // TODO: use this everywhere needed
	if GetEnv(ctx) == Prod {
		doIfProd()
	}
}
