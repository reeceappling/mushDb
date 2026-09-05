package env

import "context"

type envKey string

const envCtxKey envKey = "environment"
const Prod = "prod"
const Cert = "cert"
const Qual = "qual"
const Dev = "dev"

func GetEnv(ctx context.Context) string {
	if out, exists := ctx.Value(envCtxKey).(string); exists {
		return out
	}
	return Prod
}
func SetEnv(ctx context.Context, env string) context.Context {
	return context.WithValue(ctx, envCtxKey, env)
}
func IfDev(ctx context.Context, doIfDev func() error) error {
	e := GetEnv(ctx)
	if e == Dev || e == Qual {
		return doIfDev()
	}
	return nil
}
func IfCert(ctx context.Context, doIfProd func() error) error {
	if GetEnv(ctx) == Cert {
		return doIfProd()
	}
	return nil
}
func IfProd(ctx context.Context, doIfProd func() error) error {
	if GetEnv(ctx) == Prod {
		return doIfProd()
	}
	return nil
}
func IfNotProd(ctx context.Context, doIfProd func() error) error {
	if GetEnv(ctx) != Prod {
		return doIfProd()
	}
	return nil
}
func LogIfDev(ctx context.Context, toLog string) {
	_ = IfDev(ctx, func() error {
		println(toLog)
		return nil
	})
}
func LogAlways(toLog string) {
	println(toLog)
}
