package testutil

import (
	"context"
	"strconv"
	"sync"

	"github.com/gov4git/lib4git/git"
)

func NewCtx() context.Context {
	ctxWithAuth := git.WithAuth(context.Background(), nil)
	return context.WithValue(ctxWithAuth, counterKey{}, &counter{})
}

type counterKey struct{}

func UniqueString(ctx context.Context) string {
	return ctx.Value(counterKey{}).(*counter).Take()
}

type counter struct {
	sync.Mutex
	Next int64
}

func (x *counter) Get() int64 {
	x.Lock()
	defer x.Unlock()
	r := x.Next
	x.Next++
	return r
}

func (x *counter) Take() string {
	return strconv.Itoa(int(x.Get()))
}
