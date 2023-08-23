package git

import (
	"context"

	"github.com/gov4git/lib4git/must"
)

type contextKeyGitProxy struct{}

func WithProxy(ctx context.Context, x Proxy) context.Context {
	return context.WithValue(ctx, contextKeyGitProxy{}, x)
}

func WithCache(ctx context.Context, dir string) context.Context {
	return WithProxy(ctx, NewCache(ctx, dir))
}

func WithoutCache(ctx context.Context) context.Context {
	return WithProxy(ctx, NoCache{})
}

func getProxy(ctx context.Context) Proxy {
	x, _ := ctx.Value(contextKeyGitProxy{}).(Proxy)
	if x == nil {
		x = NoCache{}
	}
	return x
}

func CloneOne(ctx context.Context, addr Address) Cloned {
	return getProxy(ctx).CloneOne(ctx, addr)
}

func CloneAll(ctx context.Context, addr Address) Cloned {
	return getProxy(ctx).CloneAll(ctx, addr)
}

func TryCloneOne(ctx context.Context, addr Address) (cloned Cloned, err error) {
	return must.Try1(func() Cloned { return CloneOne(ctx, addr) })
}
