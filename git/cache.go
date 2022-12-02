package git

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gofrs/flock"
	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/must"
)

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

func (x *Cache) Clone(ctx context.Context, addr Address) Cloned {

	// lock access to url cache
	x.lockURL(addr.Repo)
	defer x.unlockURL(addr.Repo)

	c := &clonedFromCache{
		cache:    x,
		addr:     addr,
		diskRepo: openOrInitDisk(ctx, x.urlCachePath(addr.Repo)),
		memRepo:  openOrInitMemory(ctx),
	}
	c.pull(ctx)
	return c
}

func openOrInitDisk(ctx context.Context, path URL) *Repository {
	repo, err := git.PlainOpen(string(path))
	if err == nil {
		return repo
	}
	must.Assertf(ctx, err == git.ErrRepositoryNotExists, "%v", err)
	return InitPlain(ctx, string(path), true)
}

func openOrInitMemory(ctx context.Context) *Repository {
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	must.NoError(ctx, err)
	return repo
}

type clonedFromCache struct {
	cache    *Cache
	addr     Address
	diskRepo *Repository
	memRepo  *Repository
}

func (x *clonedFromCache) Repo() *Repository {
	return x.memRepo
}

func (x *clonedFromCache) Tree() *Tree {
	t, _ := x.memRepo.Worktree()
	return t
}

func (x *clonedFromCache) cachePath() URL {
	return x.cache.urlCachePath(x.addr.Repo)
}

func (x *clonedFromCache) Push(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.push(ctx)
}

func (x *clonedFromCache) push(ctx context.Context) {
	// push memory to disk
	PushMirror(ctx, x.memRepo, x.cachePath())
	// push disk to net
	PushMirror(ctx, x.diskRepo, x.addr.Repo)
}

func (x *clonedFromCache) Pull(ctx context.Context) {
	// lock access to url cache
	x.cache.lockURL(x.addr.Repo)
	defer x.cache.unlockURL(x.addr.Repo)
	x.pull(ctx)
}

func (x *clonedFromCache) pull(ctx context.Context) {
	// pull net into disk
	PullMirror(ctx, x.diskRepo, x.addr.Repo)
	// pull disk into memory
	PullMirror(ctx, x.memRepo, URL(x.cachePath()))
}
