package request

import "golang.org/x/net/context"

const Path string = "request-path"
const Id = "request.id"

type ctxKey string

const path = ctxKey(Path)

func GetPath(ctx context.Context) *string {
	requestPath, ok := ctx.Value(path).(string)
	if !ok {
		return nil
	}
	return &requestPath
}

func SetPath(ctx context.Context, requestPath string) context.Context {
	return context.WithValue(ctx, path, requestPath)
}
