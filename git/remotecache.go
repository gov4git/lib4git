package git

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gofrs/flock"
	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/must"
)

type RemoteCache struct {
	Dir string
	lk  sync.Mutex
	ulk map[URL]*sync.Mutex // URL locks
}

func NewRemoteCache(ctx context.Context, dir string) Proxy {
	must.NoError(ctx, os.MkdirAll(dir, 0755))
	flk := flock.New(filepath.Join(dir, "lock"))
	_, err := flk.TryLock()
	must.NoError(ctx, err)
	x := &RemoteCache{Dir: dir, ulk: map[URL]*sync.Mutex{}}
	return x
}

func (x *RemoteCache) urlLock(u URL) *sync.Mutex {
	x.lk.Lock()
	defer x.lk.Unlock()
	lk, ok := x.ulk[u]
	if !ok {
		lk = &sync.Mutex{}
		x.ulk[u] = lk
	}
	return lk
}

func (x *RemoteCache) lockURL(u URL) {
	x.urlLock(u).Lock()
}

func (x *RemoteCache) unlockURL(u URL) {
	x.urlLock(u).Unlock()
}

func (x *RemoteCache) urlCachePath(u URL) URL {
	return URL(filepath.Join(x.Dir, form.StringHashForFilename(string(u))))
}

func (x *RemoteCache) Clone(ctx context.Context, addr Address) Cloned {

	// lock access to url cache
	x.lockURL(addr.Repo)
	defer x.unlockURL(addr.Repo)

	c := &clonedRemoteCache{
		cache:    x,
		addr:     addr,
		diskRepo: openOrInitOnDisk(ctx, x.urlCachePath(addr.Repo)),
		memRepo:  initInMemory(ctx),
	}
	c.pull(ctx)

	// switch to or create branch
	err := must.Try(func() { Checkout(ctx, Worktree(ctx, c.memRepo), addr.Branch) })
	switch {
	case err == plumbing.ErrReferenceNotFound:
		must.NoError(ctx, c.memRepo.CreateBranch(&config.Branch{Name: string(addr.Branch)}))
		Checkout(ctx, Worktree(ctx, c.memRepo), addr.Branch)
	case err != nil:
		must.NoError(ctx, err)
	}

	return c
}

type clonedRemoteCache struct {
	cache    *RemoteCache
	addr     Address
	diskRepo *Repository
	memRepo  *Repository
}

func (x *clonedRemoteCache) Repo() *Repository {
	return x.memRepo
}

func (x *clonedRemoteCache) Tree() *Tree {
	t, _ := x.memRepo.Worktree()
	return t
}

func (x *clonedRemoteCache) cachePath() URL {
	return x.cache.urlCachePath(x.addr.Repo)
}

func (x *clonedRemoteCache) Push(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.push(ctx)
}

func (x *clonedRemoteCache) push(ctx context.Context) {
	// push memory to disk
	PushMirror(ctx, x.memRepo, x.cachePath())
	// push disk to net
	PushMirror(ctx, x.diskRepo, x.addr.Repo)
}

func (x *clonedRemoteCache) Pull(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.pull(ctx)
}

func (x *clonedRemoteCache) pull(ctx context.Context) {
	// pull net into disk
	PullMirror(ctx, x.diskRepo, x.addr.Repo)
	// pull disk into memory
	PullMirror(ctx, x.memRepo, URL(x.cachePath()))
}
