package git

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/gov4git/lib4git/must"
)

func nonceName() string {
	return "nonce-" + strconv.FormatUint(uint64(rand.Int63()), 36)
}

const (
	mirrorBranchesRefSpec = "refs/heads/*:refs/heads/*"
	mirrorRemotesRefSpec  = "refs/remotes/*:refs/remotes/*"
	mirrorTagsRefSpec     = "refs/tags/*:refs/tags/*"
	mirrorHeadRefSpec     = "refs/HEAD:refs/HEAD"
)

var mirrorRefSpecs = []config.RefSpec{
	mirrorBranchesRefSpec,
	// mirrorRemotesRefSpec,
	// mirrorTagsRefSpec, mirrorHeadRefSpec,
}

func branchRefSpec(b Branch) []config.RefSpec {
	return []config.RefSpec{
		config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", b, b)),
	}
}

func branchSubtreeRefSpec(b Branch) []config.RefSpec {
	return []config.RefSpec{
		config.RefSpec(fmt.Sprintf("refs/heads/%s/*:refs/heads/%s/*", b, b)),
	}
}

func PushAll(ctx context.Context, repo *Repository, to URL) {
	Push(ctx, repo, to, mirrorRefSpecs)
}

func PullAll(ctx context.Context, repo *Repository, from URL) {
	Pull(ctx, repo, from, mirrorRefSpecs)
}

func overwriteRemote(ctx context.Context, repo *Repository, to URL, refspecs []config.RefSpec) (*git.Remote, string) {
	remoteName := to.Hash()
	remoteConfig := &config.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{string(to)},
		Fetch: refspecs,
	}
	remote, err := repo.CreateRemote(remoteConfig)
	if err != nil {
		if err != git.ErrRemoteExists {
			must.NoError(ctx, err)
		}
		err = repo.DeleteRemote(remoteName)
		must.NoError(ctx, err)
		remote, err = repo.CreateRemote(remoteConfig)
		must.NoError(ctx, err)
	}
	return remote, remoteName
}

// Push implements `git push`. It creates a remote, whose name is the hash of the remote repo's URL.
// If the remote exists, it is overwritten.
func Push(ctx context.Context, repo *Repository, to URL, refspecs []config.RefSpec) {
	remote, remoteName := overwriteRemote(ctx, repo, to, refspecs)
	must.NoError(ctx, remote.PushContext(ctx, &git.PushOptions{
		RemoteName: remoteName,
		Auth:       GetAuth(ctx, to),
	}))
}

// Push implements `git pull`. It creates a remote, whose name is the hash of the remote repo's URL.
// If the remote exists, it is overwritten.
func Pull(ctx context.Context, repo *Repository, from URL, refspecs []config.RefSpec) {
	remote, remoteName := overwriteRemote(ctx, repo, from, refspecs)
	err := remote.FetchContext(ctx, &git.FetchOptions{
		RemoteName: remoteName,
		Auth:       GetAuth(ctx, from),
		Force:      true,
	})
	must.Assertf(ctx,
		err == transport.ErrEmptyRemoteRepository ||
			err == git.NoErrAlreadyUpToDate ||
			err == nil, "%v", err)
}

// PushOnce implements `git push` without creating a new remote entry.
func PushOnce(ctx context.Context, repo *Repository, to URL, refspecs []config.RefSpec) {
	nonce := nonceName()
	remote := git.NewRemote(
		repo.Storer,
		&config.RemoteConfig{
			Name:  nonce,
			URLs:  []string{string(to)},
			Fetch: refspecs,
		},
	)
	err := remote.PushContext(ctx, &git.PushOptions{
		RemoteName: nonce,
		Auth:       GetAuth(ctx, to),
	})
	if IsAlreadyUpToDate(err) {
		return
	}
	must.NoError(ctx, err)
}

// PullOnce implements `git pull` without creating a new remote entry.
// Panics with authentication required, i/o timeout, repository not found.
func PullOnce(ctx context.Context, repo *Repository, from URL, refspecs []config.RefSpec) {
	nonce := nonceName()
	remote := git.NewRemote(
		repo.Storer,
		&config.RemoteConfig{
			Name:  nonce,
			URLs:  []string{string(from)},
			Fetch: refspecs,
		},
	)
	err := remote.FetchContext(ctx, &git.FetchOptions{
		RemoteName: nonce,
		Auth:       GetAuth(ctx, from),
		Force:      true,
	})
	// panic on authentication required, i/o timeout, repository not found (repo is inaccessible)
	// ignore empty repo, already up to date, branch not found
	must.Assertf(ctx,
		err == nil ||
			IsRemoteRepoIsEmpty(err) ||
			IsAlreadyUpToDate(err) ||
			IsNoMatchingRefSpec(err), "%v", err)
}
