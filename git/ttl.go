package git

import (
	"context"
	"sync"
	"time"
)

// ttl manager in the context

type contextKeyTTLManager struct{}

func WithTTL(ctx context.Context, am *TTLManager) context.Context {
	if am == nil {
		am = NewTTLManager()
	}
	return context.WithValue(ctx, contextKeyTTLManager{}, am)
}

func SetTTL(ctx context.Context, forRepo URL, a time.Duration) {
	ctx.Value(contextKeyTTLManager{}).(*TTLManager).SetTTL(forRepo, a)
}

func GetTTL(ctx context.Context, forRepo URL) time.Duration {
	if tm, ok := ctx.Value(contextKeyTTLManager{}).(*TTLManager); ok {
		return tm.GetTTL(forRepo)
	}
	return 0
}

// TTLManager provides TTL hints given a repo URL.
type TTLManager struct {
	lk  sync.Mutex
	ttl map[URL]time.Duration
}

func NewTTLManager() *TTLManager {
	return &TTLManager{ttl: map[URL]time.Duration{}}
}

func (x *TTLManager) SetTTL(forRepo URL, a time.Duration) {
	x.lk.Lock()
	defer x.lk.Unlock()
	x.ttl[forRepo] = a
}

func (x *TTLManager) GetTTL(forRepo URL) time.Duration {
	x.lk.Lock()
	defer x.lk.Unlock()
	return x.ttl[forRepo]
}
