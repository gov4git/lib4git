package git

import (
	"context"
	"path/filepath"

	"github.com/gov4git/lib4git/base"
)

type NoCache struct {
	dir string // if non empty, repositories will be created on disk within this dir
}

func NewNoCacheInMemory() Proxy {
	return NoCache{}
}

func NewNoCacheOnDisk(dir string) Proxy {
	return NoCache{dir: dir}
}

func (x NoCache) CloneOne(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, false)
}

func (x NoCache) CloneAll(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, true)
}

func (x NoCache) clone(ctx context.Context, addr Address, all bool) Cloned {
	var repo *Repository
	if x.dir == "" {
		repo = initInMemory(ctx)
	} else {
		p := URL(filepath.Join(x.dir, nonceName()))
		base.Infof("materializing repo %v on disk %v\n", addr.Repo, p)
		repo = openOrInitOnDisk(ctx, p, false)
	}
	c := &clonedNoCache{all: all, addr: addr, repo: repo}
	c.Pull(ctx)
	switchToBranch(ctx, c.repo, addr.Branch)
	return c
}

type clonedNoCache struct {
	all  bool
	addr Address
	repo *Repository
}

func (x *clonedNoCache) Push(ctx context.Context) {
	PushOnce(ctx, x.repo, x.addr.Repo, mirrorRefSpecs)
}

func (x *clonedNoCache) Pull(ctx context.Context) {
	PullOnce(ctx, x.repo, x.addr.Repo, clonePullRefSpecs(x.addr, x.all))
}

func (x *clonedNoCache) Repo() *Repository {
	return x.repo
}

func (x *clonedNoCache) Tree() *Tree {
	t, _ := x.Repo().Worktree()
	return t
}
