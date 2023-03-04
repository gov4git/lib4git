package git

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gofrs/flock"
	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/must"
)

// Cache is a Proxy.
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

func (x *Cache) urlCachePath(u URL) URL {
	return URL(filepath.Join(x.Dir, form.StringHashForFilename(string(u))))
}

func (x *Cache) CloneOne(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, false)
}

func (x *Cache) CloneAll(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, true)
}

func (x *Cache) clone(ctx context.Context, addr Address, all bool) Cloned {

	// lock access to url cache
	x.lockURL(addr.Repo)
	defer x.unlockURL(addr.Repo)

	c := &clonedCacheProxy{
		all:      all,
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
		SetHeadToBranch(ctx, c.memRepo, addr.Branch)
		// must.NoError(ctx, Worktree(ctx, c.memRepo).Reset(&git.ResetOptions{Mode: git.HardReset}))
	case err != nil:
		must.NoError(ctx, err)
	}

	return c
}

func openOrInitOnDisk(ctx context.Context, path URL) *Repository {
	repo, err := git.PlainOpen(string(path))
	if err == nil {
		return repo
	}
	must.Assertf(ctx, err == git.ErrRepositoryNotExists, "%v", err)
	return InitPlain(ctx, string(path), true) // cache must be bare, otherwise checkout branch cannot be pushed
}

type clonedCacheProxy struct {
	cache    *Cache
	all      bool
	addr     Address
	diskRepo *Repository
	memRepo  *Repository
}

func (x *clonedCacheProxy) Repo() *Repository {
	return x.memRepo
}

func (x *clonedCacheProxy) Tree() *Tree {
	t, _ := x.memRepo.Worktree()
	return t
}

func (x *clonedCacheProxy) cachePath() URL {
	return x.cache.urlCachePath(x.addr.Repo)
}

func (x *clonedCacheProxy) Push(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.push(ctx)
}

// push pushes all local branches to the remote origin.
func (x *clonedCacheProxy) push(ctx context.Context) {
	PushOnce(ctx, x.memRepo, x.cachePath(), mirrorRefSpecs) // push memory to disk
	PushOnce(ctx, x.diskRepo, x.addr.Repo, mirrorRefSpecs)  // push disk to net
}

func (x *clonedCacheProxy) Pull(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.pull(ctx)
}

// pull pulls only the branch explicitly named in the clone invocation.
func (x *clonedCacheProxy) pull(ctx context.Context) {
	refSpec := clonePullRefSpecs(x.addr, x.all)
	PullOnce(ctx, x.diskRepo, x.addr.Repo, refSpec)  // pull net into disk
	PullOnce(ctx, x.memRepo, x.cachePath(), refSpec) // pull disk into memory
}

func clonePullRefSpecs(addr Address, all bool) []config.RefSpec {
	if all {
		return mirrorRefSpecs
	}
	return branchRefSpec(addr.Branch)
}
