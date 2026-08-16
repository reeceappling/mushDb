package request

import (
	"github.com/google/uuid"
	"github.com/reeceappling/mushDb/api/request/unix"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

const Path string = "request-path"
const Id = "request.id"

type ctxKey string

const nowKey ctxKey = "request.now.unix"

const path = ctxKey(Path)
const id = ctxKey(Id)

func GetPath(ctx context.Context) *string { // TODO: USE!
	requestPath, ok := ctx.Value(path).(string)
	if !ok {
		return nil
	}
	return &requestPath
}

func SetPath(ctx context.Context, requestPath string) context.Context {
	return context.WithValue(ctx, path, requestPath)
}
func GetId(ctx context.Context) *string { // TODO: USE!
	idFromCtx, ok := ctx.Value(id).(string)
	if !ok {
		return nil
	}
	return &idFromCtx
}

func WithId(ctx context.Context, optionalId *string) context.Context {
	var newId string
	if optionalId != nil {
		newId = *optionalId
	} else {
		newId = uuid.New().String()
	}
	return context.WithValue(ctx, id, newId)
}

// UnixTime grabs the unixTime from the context if it is set,
// but otherwise calculates it and sets it on the context, also returning the time.
func UnixTime(ctx context.Context) (context.Context, unix.Time) {
	t, ok := ctx.Value(nowKey).(unix.Time)
	if !ok {
		t = unix.TimeForNow()
		return context.WithValue(ctx, nowKey, t), t
	}
	return ctx, t
}

// UnixTimeInTxn Gets the current unix time, sets it on the context, grabs a session from the context,
// and returns a new session context containing the unixTime so it does not need to be recalculated later
func UnixTimeInTxn(ctx mongo.SessionContext) (mongo.SessionContext, unix.Time) {
	sess := mongo.SessionFromContext(ctx)
	ctxTemp, t := UnixTime(ctx)
	sessCtx := mongo.NewSessionContext(ctxTemp, sess)
	return sessCtx, t
}
