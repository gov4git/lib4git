package git

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/gov4git/lib4git/must"
)

type Proxy interface {
	Clone(ctx context.Context, addr Address) Cloned
	ClonePrefix(ctx context.Context, prefix Address) Cloned // clone all branches with prefix addr.Branch
}

type Cloned interface {
	Push(context.Context)
	Pull(context.Context)
	Repo() *Repository
	Tree() *Tree
}

func GitClone(ctx context.Context, addr Address) Cloned {
	return &clonedBranch{prefix: false, addr: addr, repo: cloneToMemory(ctx, addr)}
}

func GitClonePrefix(ctx context.Context, addr Address) Cloned {
	return &clonedBranch{prefix: true, addr: addr, repo: cloneToMemory(ctx, addr)}
}

func GitCloneOrInit(ctx context.Context, addr Address) Cloned {
	repo, _ := cloneOrInit(ctx, addr)
	return &clonedBranch{prefix: false, addr: addr, repo: repo}
}

type clonedBranch struct {
	prefix bool
	addr   Address
	repo   *Repository
}

func (x *clonedBranch) Push(ctx context.Context) {
	if err := x.repo.PushContext(ctx, &git.PushOptions{
		RefSpecs: mirrorRefSpecs,
		Auth:     GetAuth(ctx, x.addr.Repo),
	}); err != nil {
		must.Panic(ctx, err)
	}
}

func (x *clonedBranch) Pull(ctx context.Context) {
	if err := x.repo.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: clonePullRefSpecs(x.addr, x.prefix),
		Auth:     GetAuth(ctx, x.addr.Repo),
	}); err != nil {
		must.Panic(ctx, err)
	}
}

func (x *clonedBranch) Repo() *Repository {
	return x.repo
}

func (x *clonedBranch) Tree() *Tree {
	t, _ := x.Repo().Worktree()
	return t
}
