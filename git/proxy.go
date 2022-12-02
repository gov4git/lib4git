package git

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/gov4git/lib4git/must"
)

type Proxy interface {
	Clone(ctx context.Context, addr Address) Cloned
}

type Cloned interface {
	Push(context.Context)
	Pull(context.Context)
	Repo() *Repository
	Tree() *Tree
}

func ClonedTree(c Cloned) *Tree {
	_, t := ClonedRepoTree(c)
	return t
}

func ClonedRepoTree(c Cloned) (*Repository, *Tree) {
	t, err := c.Repo().Worktree()
	if err != nil {
		panic("clone without a tree")
	}
	return c.Repo(), t
}

func Clone(ctx context.Context, addr Address) Cloned {
	return &clonedBranch{addr: addr, repo: cloneRepo(ctx, addr)}
}

func CloneOrInit(ctx context.Context, addr Address) Cloned {
	repo, _ := cloneOrInit(ctx, addr)
	return &clonedBranch{addr: addr, repo: repo}
}

type clonedBranch struct {
	addr Address
	repo *Repository
}

func (x *clonedBranch) Push(ctx context.Context) {
	if err := x.repo.PushContext(ctx, &git.PushOptions{
		Auth: GetAuth(ctx, x.addr.Repo),
	}); err != nil {
		must.Panic(ctx, err)
	}
}

func (x *clonedBranch) Pull(ctx context.Context) {
	if err := x.repo.FetchContext(ctx, &git.FetchOptions{
		Auth: GetAuth(ctx, x.addr.Repo),
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
