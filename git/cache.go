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
	return x.clone(ctx, addr, x.makeRepo(ctx), false)
}

func (x *Cache) CloneOneTo(ctx context.Context, addr Address, to *Repository) Cloned {
	return x.clone(ctx, addr, to, false)
}

func (x *Cache) CloneAll(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, x.makeRepo(ctx), true)
}

func (x *Cache) CloneAllTo(ctx context.Context, addr Address, to *Repository) Cloned {
	return x.clone(ctx, addr, to, true)
}

func (x *Cache) makeRepo(ctx context.Context) *Repository {
	return initInMemory(ctx)
}

func (x *Cache) clone(ctx context.Context, addr Address, to *Repository, all bool) Cloned {

	// lock access to url cache
	x.lockURL(addr.Repo)
	defer x.unlockURL(addr.Repo)

	c := &clonedCacheProxy{
		all:       all,
		cache:     x,
		addr:      addr,
		cacheRepo: openOrInitOnDisk(ctx, x.urlCachePath(addr.Repo), true), // cache must be bare, otherwise checkout branch cannot be pushed
		userRepo:  to,
	}
	c.pull(ctx)
	switchToBranch(ctx, c.userRepo, addr.Branch)
	return c
}

func switchToBranch(ctx context.Context, repo *Repository, branch Branch) {
	err := must.Try(func() { Checkout(ctx, Worktree(ctx, repo), branch) })
	switch {
	case err == plumbing.ErrReferenceNotFound:
		must.NoError(ctx, repo.CreateBranch(&config.Branch{Name: string(branch)}))
		SetHeadToBranch(ctx, repo, branch)
		// must.NoError(ctx, Worktree(ctx, repo).Reset(&git.ResetOptions{Mode: git.HardReset}))
	case err != nil:
		must.NoError(ctx, err)
	}
}

func openOrInitOnDisk(ctx context.Context, path URL, bare bool) *Repository {
	repo, err := git.PlainOpen(string(path))
	if err == nil {
		return repo
	}
	must.Assertf(ctx, err == git.ErrRepositoryNotExists, "%v", err)
	return InitPlain(ctx, string(path), bare)
}

type clonedCacheProxy struct {
	cache     *Cache
	all       bool
	addr      Address
	cacheRepo *Repository
	userRepo  *Repository
}

func (x *clonedCacheProxy) Repo() *Repository {
	return x.userRepo
}

func (x *clonedCacheProxy) Tree() *Tree {
	t, _ := x.userRepo.Worktree()
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
	PushOnce(ctx, x.userRepo, x.cachePath(), mirrorRefSpecs) // push memory to disk
	PushOnce(ctx, x.cacheRepo, x.addr.Repo, mirrorRefSpecs)  // push disk to net
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
	PullOnce(ctx, x.cacheRepo, x.addr.Repo, refSpec)  // pull net into disk
	PullOnce(ctx, x.userRepo, x.cachePath(), refSpec) // pull disk into memory
}

func clonePullRefSpecs(addr Address, all bool) []config.RefSpec {
	if all {
		return mirrorRefSpecs
	}
	return branchRefSpec(addr.Branch)
}
