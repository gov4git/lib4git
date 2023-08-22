package git

import (
	"context"
	"sync"

	"github.com/gov4git/lib4git/must"
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

func UseNoCacheOnDisk(ctx context.Context, dir string) {
	cacheLk.Lock()
	defer cacheLk.Unlock()
	proxy = NewNoCacheOnDisk(dir)
}

func getProxy() Proxy {
	cacheLk.Lock()
	defer cacheLk.Unlock()
	return proxy
}

func CloneOne(ctx context.Context, addr Address) Cloned {
	if pxy := getProxy(); pxy != nil {
		return pxy.CloneOne(ctx, addr)
	}
	return NoCache{}.CloneOne(ctx, addr)
}

func CloneAll(ctx context.Context, addr Address) Cloned {
	if pxy := getProxy(); pxy != nil {
		return pxy.CloneAll(ctx, addr)
	}
	return NoCache{}.CloneAll(ctx, addr)
}

func TryCloneOne(ctx context.Context, addr Address) (cloned Cloned, err error) {
	return must.Try1(func() Cloned { return CloneOne(ctx, addr) })
}
