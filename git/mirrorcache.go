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

type MirrorCache struct {
	Dir string
	lk  sync.Mutex
	ulk map[URL]*sync.Mutex // URL locks
}

func NewMirrorCache(ctx context.Context, dir string) Proxy {
	must.NoError(ctx, os.MkdirAll(dir, 0755))
	flk := flock.New(filepath.Join(dir, "lock"))
	_, err := flk.TryLock()
	must.NoError(ctx, err)
	x := &MirrorCache{Dir: dir, ulk: map[URL]*sync.Mutex{}}
	return x
}

func (x *MirrorCache) urlLock(u URL) *sync.Mutex {
	x.lk.Lock()
	defer x.lk.Unlock()
	lk, ok := x.ulk[u]
	if !ok {
		lk = &sync.Mutex{}
		x.ulk[u] = lk
	}
	return lk
}

func (x *MirrorCache) lockURL(u URL) {
	x.urlLock(u).Lock()
}

func (x *MirrorCache) unlockURL(u URL) {
	x.urlLock(u).Unlock()
}

func (x *MirrorCache) urlCachePath(u URL) URL {
	return URL(filepath.Join(x.Dir, form.StringHashForFilename(string(u))))
}

func (x *MirrorCache) Clone(ctx context.Context, addr Address) Cloned {

	// lock access to url cache
	x.lockURL(addr.Repo)
	defer x.unlockURL(addr.Repo)

	c := &clonedMirrorCache{
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

type clonedMirrorCache struct {
	cache    *MirrorCache
	addr     Address
	diskRepo *Repository
	memRepo  *Repository
}

func (x *clonedMirrorCache) Repo() *Repository {
	return x.memRepo
}

func (x *clonedMirrorCache) Tree() *Tree {
	t, _ := x.memRepo.Worktree()
	return t
}

func (x *clonedMirrorCache) cachePath() URL {
	return x.cache.urlCachePath(x.addr.Repo)
}

func (x *clonedMirrorCache) Push(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.push(ctx)
}

// push pushes all local branches to the remote origin.
func (x *clonedMirrorCache) push(ctx context.Context) {
	PushOnce(ctx, x.memRepo, x.cachePath(), mirrorRefSpecs) // push memory to disk
	PushOnce(ctx, x.diskRepo, x.addr.Repo, mirrorRefSpecs)  // push disk to net
}

func (x *clonedMirrorCache) Pull(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.pull(ctx)
}

// pull pulls only the branch explicitly named in the clone invocation.
func (x *clonedMirrorCache) pull(ctx context.Context) {
	refSpec := branchRefSpec(x.addr.Branch)
	PullOnce(ctx, x.diskRepo, x.addr.Repo, refSpec)  // pull net into disk
	PullOnce(ctx, x.memRepo, x.cachePath(), refSpec) // pull disk into memory
}
