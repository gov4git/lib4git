package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/must"
)

// replicaClone is an in-memory clone, backed by a local, on-disk copy of a remote repo used for caching.
// replicaClone are re-entrant across multiple processes, as they use file locks to guarantee correctness.
type replicaClone struct {
	cacheDir    string  // replica base directory
	address     Address // remote address
	allBranches bool
	ttl         time.Duration
	diskRepo    *Repository
	memRepo     *Repository
}

func newReplicaClone(ctx context.Context, cacheDir string, address Address, allBranches bool, ttl time.Duration) *replicaClone {
	must.NoError(ctx, os.MkdirAll(cacheDir, 0755))
	return &replicaClone{
		cacheDir:    cacheDir,
		address:     address,
		allBranches: allBranches,
		ttl:         ttl,
		diskRepo:    openOrInitOnDisk(ctx, replicaPathURL(cacheDir, address), true), // cache must be bare, otherwise checkout branch cannot be pushed,
		memRepo:     initInMemory(ctx),
	}
}

func (x *replicaClone) Repo() *Repository {
	return x.memRepo
}

func (x *replicaClone) Tree() *Tree {
	t, _ := x.memRepo.Worktree()
	return t
}

func (x *replicaClone) replicaDiskRepoURL() URL {
	return replicaPathURL(x.cacheDir, x.address)
}

func (x *replicaClone) replicaLockPath() string {
	return replicaLockPath(x.cacheDir, x.address)
}

func (x *replicaClone) replicaTimestampPath() string {
	return replicaTimestampPath(x.cacheDir, x.address)
}

// replicaPathURL returns the git URL for the local path CACHE_DIR/ADDRESS_HASH/repo
func replicaPathURL(cacheDir string, addr Address) URL {
	return URL(filepath.Join(cacheDir, form.StringHashForFilename(string(addr.String())), "repo"))
}

// replicaLockPath returns the local path CACHE_DIR/ADDRESS_HASH/lock
func replicaLockPath(cacheDir string, addr Address) string {
	return filepath.Join(cacheDir, form.StringHashForFilename(string(addr.String())), "lock")
}

// replicaTimestampPath returns the local path CACHE_DIR/ADDRESS_HASH/stamp
func replicaTimestampPath(cacheDir string, addr Address) string {
	return filepath.Join(cacheDir, form.StringHashForFilename(string(addr.String())), "stamp")
}

var ReplicaLockRetryDelay = time.Millisecond * 100

func (x *replicaClone) Push(ctx context.Context) {
	// lock on disk cache
	flk := flock.New(x.replicaLockPath())
	locked, err := flk.TryLockContext(ctx, ReplicaLockRetryDelay)
	must.NoError(ctx, err)
	must.Assertf(ctx, locked, "cache replica lock failed (%v)", err)
	defer flk.Unlock()
	// perform push
	x.push(ctx)
}

func (x *replicaClone) push(ctx context.Context) {
	x.invalidateCache(ctx)
	PushOnce(ctx, x.memRepo, x.replicaDiskRepoURL(), mirrorRefSpecs) // push memory to disk
	PushOnce(ctx, x.diskRepo, x.address.Repo, mirrorRefSpecs)        // push disk to remote
	x.validateCache(ctx)
}

func (x *replicaClone) Pull(ctx context.Context) {
	// lock on disk cache
	flk := flock.New(x.replicaLockPath())
	locked, err := flk.TryLockContext(ctx, ReplicaLockRetryDelay)
	must.NoError(ctx, err)
	must.Assertf(ctx, locked, "cache replica lock failed (%v)", err)
	defer flk.Unlock()
	// perform fetch
	x.pull(ctx)
}

func (x *replicaClone) pull(ctx context.Context) {
	if x.isCacheValid(ctx) {
		return
	}
	refSpec := clonePullRefSpecs(x.address, x.allBranches)
	PullOnce(ctx, x.diskRepo, x.address.Repo, refSpec) // pull remote into disk
	x.validateCache(ctx)
	PullOnce(ctx, x.memRepo, x.replicaDiskRepoURL(), refSpec) // pull disk into memory
}

func (x *replicaClone) isCacheValid(ctx context.Context) bool {
	fi, err := os.Stat(x.replicaTimestampPath())
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	must.NoError(ctx, err)
	return time.Now().Sub(fi.ModTime()) <= x.ttl
}

func (x *replicaClone) validateCache(ctx context.Context) {
	err := os.WriteFile(x.replicaTimestampPath(), []byte(time.Now().String()), 0644)
	must.NoError(ctx, err)
}

func (x *replicaClone) invalidateCache(ctx context.Context) {
	err := os.Remove(x.replicaTimestampPath())
	must.NoError(ctx, err)
}
