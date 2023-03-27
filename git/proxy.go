package git

import (
	"context"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gov4git/lib4git/must"
)

type Proxy interface {
	// CloneOne fetches just the branch specified in addr then checks out the branch in addr, creating it if not present.
	CloneOne(ctx context.Context, addr Address) Cloned
	// CloneAll fetches all branches then checks out the branch in addr, creating it if not present.
	CloneAll(ctx context.Context, addr Address) Cloned
}

type Cloned interface {
	// Push all branches to the origin.
	Push(context.Context)
	// Pull the branches indicated by the clone call that created this clone.
	Pull(context.Context)
	Repo() *Repository
	Tree() *Tree
}

func cloneOneNoProxy(ctx context.Context, addr Address) Cloned {
	return &clonedNoProxy{all: false, addr: addr, repo: cloneToMemoryOrInit(ctx, addr)}
}

func cloneAllNoProxy(ctx context.Context, addr Address) Cloned {
	c := &clonedNoProxy{all: true, addr: addr, repo: cloneToMemoryOrInit(ctx, addr)}
	c.Pull(ctx)
	return c
}

func cloneToMemoryOrInit(ctx context.Context, addr Address) *Repository {
	repo, err := must.Try1(func() *Repository { return cloneToMemory(ctx, addr) })
	if err == nil {
		return repo
	}
	_, isNoBranch := err.(git.NoMatchingRefSpecError)
	if !isNoBranch && err != transport.ErrEmptyRemoteRepository && err != plumbing.ErrReferenceNotFound {
		must.Panic(ctx, err)
	}
	repo = initInMemory(ctx)

	_, err = repo.CreateRemote(&config.RemoteConfig{Name: Origin, URLs: []string{string(addr.Repo)}})
	must.NoError(ctx, err)

	err = repo.CreateBranch(&config.Branch{Name: string(addr.Branch), Remote: Origin})
	must.NoError(ctx, err)

	ChangeDefaultBranch(ctx, repo, addr.Branch)

	return repo
}

func initInMemory(ctx context.Context) *Repository {
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	must.NoError(ctx, err)
	return repo
}

func cloneToMemory(ctx context.Context, addr Address) *Repository {
	repo, err := git.CloneContext(ctx,
		memory.NewStorage(),
		memfs.New(),
		&git.CloneOptions{
			URL:           string(addr.Repo),
			Auth:          GetAuth(ctx, addr.Repo),
			ReferenceName: plumbing.NewBranchReferenceName(string(addr.Branch)),
		},
	)
	must.NoError(ctx, err)

	return repo
}

type clonedNoProxy struct {
	all  bool
	addr Address
	repo *Repository
}

func (x *clonedNoProxy) Push(ctx context.Context) {
	if err := x.repo.PushContext(ctx, &git.PushOptions{
		RefSpecs: mirrorRefSpecs,
		Auth:     GetAuth(ctx, x.addr.Repo),
	}); err != nil {
		must.Panic(ctx, err)
	}
}

func (x *clonedNoProxy) Pull(ctx context.Context) {
	err := x.repo.FetchContext(ctx, &git.FetchOptions{
		RefSpecs: clonePullRefSpecs(x.addr, x.all),
		Auth:     GetAuth(ctx, x.addr.Repo),
		Force:    true,
	})
	if err == transport.ErrEmptyRemoteRepository {
		return
	}
	must.NoError(ctx, err)
}

func (x *clonedNoProxy) Repo() *Repository {
	return x.repo
}

func (x *clonedNoProxy) Tree() *Tree {
	t, _ := x.Repo().Worktree()
	return t
}
