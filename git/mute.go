package git

import "context"

type muteStagingCtxKey struct{}

func MuteStaging(ctx context.Context) context.Context {
	return context.WithValue(ctx, muteStagingCtxKey{}, true)
}

func UnmuteStaging(ctx context.Context) context.Context {
	return context.WithValue(ctx, muteStagingCtxKey{}, false)
}

func IsStagingMuted(ctx context.Context) bool {
	v, ok := ctx.Value(muteStagingCtxKey{}).(bool)
	return ok && v
}
