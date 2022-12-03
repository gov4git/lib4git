package git

import (
	"context"
	"sync"
)

var (
	cacheLk sync.Mutex
	proxy   Proxy
)

func UseCache(ctx context.Context, dir string) {
	cacheLk.Lock()
	defer cacheLk.Unlock()
	proxy = NewCache(ctx, dir)
}

func getProxy() Proxy {
	cacheLk.Lock()
	defer cacheLk.Unlock()
	return proxy
}

func Clone(ctx context.Context, addr Address) Cloned {
	if pxy := getProxy(); pxy != nil {
		return pxy.Clone(ctx, addr)
	}
	return GitClone(ctx, addr)
}

func CloneOrInit(ctx context.Context, addr Address) Cloned {
	if pxy := getProxy(); pxy != nil {
		return pxy.Clone(ctx, addr)
	}
	return GitCloneOrInit(ctx, addr)
}
