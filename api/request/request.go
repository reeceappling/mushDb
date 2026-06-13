package request

import (
	"github.com/reeceappling/mushDb/api/request/unix"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

const Path string = "request-path"
const Id = "request.id"

type ctxKey string

const path = ctxKey(Path)
const nowKey ctxKey = "request.now.unix"

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

func UnixTime(ctx context.Context) (context.Context, unix.Time) {
	t, ok := ctx.Value(nowKey).(unix.Time)
	if !ok {
		t = unix.TimeForNow()
		return context.WithValue(ctx, nowKey, t), t
	}
	return ctx, t
}

func UnixTimeInTxn(ctx mongo.SessionContext) (mongo.SessionContext, unix.Time) {
	ctxTemp, t := UnixTime(ctx)
	sess := mongo.SessionFromContext(ctx)
	return mongo.NewSessionContext(ctxTemp, sess), t
}
