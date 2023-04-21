package git

import (
	"context"
	"path/filepath"
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
	return x.clone(ctx, addr, x.makeRepo(ctx), false)
}

func (x NoCache) CloneOneTo(ctx context.Context, addr Address, to *Repository) Cloned {
	return x.clone(ctx, addr, to, false)
}

func (x NoCache) CloneAll(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, x.makeRepo(ctx), true)
}

func (x NoCache) CloneAllTo(ctx context.Context, addr Address, to *Repository) Cloned {
	return x.clone(ctx, addr, to, true)
}

func (x NoCache) makeRepo(ctx context.Context) *Repository {
	if x.dir == "" {
		return initInMemory(ctx)
	} else {
		p := URL(filepath.Join(x.dir, nonceName()))
		return openOrInitOnDisk(ctx, p, false)
	}
}

func (x NoCache) clone(ctx context.Context, addr Address, to *Repository, all bool) Cloned {
	c := &clonedNoCache{all: all, addr: addr, repo: to}
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
