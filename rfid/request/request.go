package request

import "golang.org/x/net/context"

const Path = "request-path"
const Id = "request.id"

func GetPath(ctx context.Context) *string {
	requestPath, ok := ctx.Value(Path).(string)
	if !ok {
		return nil
	}
	return &requestPath
}

func SetPath(ctx context.Context, requestPath string) context.Context {
	return context.WithValue(ctx, Path, requestPath)
}
