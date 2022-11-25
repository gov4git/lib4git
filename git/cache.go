package git

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5/config"
	"github.com/gofrs/flock"
	"github.com/gov4git/lib4git/must"
)

type Proxy interface {
	Clone(ctx context.Context, addr Address, refspecs []config.RefSpec) Cloned
}

type Cloned interface {
	Push(context.Context)
	Pull(context.Context)
	Repo() *Repository
}

type Cache struct {
	Dir string
	lk  sync.Mutex
	ulk map[URL]*sync.Mutex // URL locks
}

func NewCache(ctx context.Context, dir string) Proxy {
	must.NoError(ctx, os.MkdirAll(dir, 0755))
	flk := flock.New(filepath.Join(dir, "lock"))
	_, err := flk.TryLock()
	must.NoError(ctx, err)
	x := &Cache{Dir: dir, ulk: map[URL]*sync.Mutex{}}
	return x
}

func (x *Cache) urlLock(u URL) *sync.Mutex {
	x.lk.Lock()
	defer x.lk.Unlock()
	lk, ok := x.ulk[u]
	if !ok {
		lk = &sync.Mutex{}
		x.ulk[u] = lk
	}
	return lk
}

func (x *Cache) lockURL(u URL) {
	x.urlLock(u).Lock()
}

func (x *Cache) unlockURL(u URL) {
	x.urlLock(u).Unlock()
}

func (x *Cache) Clone(ctx context.Context, addr Address, refspecs []config.RefSpec) Cloned {
	x.lockURL(addr.Repo)
	defer x.unlockURL(addr.Repo)
	XXX
	return &clonedFromCache{addr: addr, refspecs: refspecs, repo: repo}
}

type clonedFromCache struct {
	addr     Address
	refspecs []config.RefSpec
	repo     *Repository
}

func (x *clonedFromCache) Repo() *Repository {
	return x.repo
}

func (x *clonedFromCache) Push(ctx context.Context) {
	XXX
}

func (x *clonedFromCache) Pull(ctx context.Context) {
	XXX
}

// func pushThruCache(
// 	ctx context.Context,
// 	local *Repository,
// 	cache *Repository,
// 	origin URL,
// 	refspecs []config.RefSpec,
// ) {
// 	XXX
// }
