package git

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gov4git/lib4git/must"
)

type Cache struct {
	cacheDir string // root of all replicas
}

func NewCache(_ context.Context, dir string) *Cache {
	return &Cache{cacheDir: dir}
}

func (x *Cache) CloneOne(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, false)
}

func (x *Cache) CloneAll(ctx context.Context, addr Address) Cloned {
	return x.clone(ctx, addr, true)
}

func (x *Cache) clone(ctx context.Context, addr Address, all bool) Cloned {
	c := newReplicaClone(ctx, x.cacheDir, addr, all, GetTTL(ctx, addr.Repo))
	c.pull(ctx)
	switchToBranch(ctx, c.memRepo, addr.Branch)
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

func OpenOrInitOnDisk(ctx context.Context, path URL, bare bool) *Repository {
	repo, err := git.PlainOpen(string(path))
	if err == nil {
		return repo
	}
	must.Assertf(ctx, err == git.ErrRepositoryNotExists, "%v", err)
	return InitPlain(ctx, string(path), bare)
}

func clonePullRefSpecs(addr Address, all bool) []config.RefSpec {
	if all {
		return mirrorRefSpecs
	}
	return branchRefSpec(addr.Branch)
}
