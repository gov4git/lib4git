package testutil

import (
	"context"
	"math/rand"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/gov4git/lib4git/git"
)

func NewCtx(t *testing.T, useCache bool) context.Context {
	ctx := git.WithAuth(context.Background(), nil)
	ctx = git.WithTTL(ctx, nil)
	ctx = context.WithValue(ctx, counterKey{}, &counter{})
	if useCache {
		return git.WithCache(ctx, filepath.Join(t.TempDir(), "cache-"+strconv.FormatUint(uint64(rand.Int63()), 36)))
	} else {
		return ctx
	}
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
