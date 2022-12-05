package git

import (
	"context"
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

const mirrorBranchesRefSpec = "refs/heads/*:refs/heads/*"
const mirrorTagsRefSpec = "refs/tags/*:refs/tags/*"
const mirrorHeadRefSpec = "refs/HEAD:refs/HEAD"

var mirrorRefSpecs = []config.RefSpec{mirrorBranchesRefSpec, mirrorTagsRefSpec /*, mirrorHeadRefSpec*/}

func PushAll(ctx context.Context, repo *Repository, to URL) {
	Push(ctx, repo, to, mirrorRefSpecs)
}

func PullAll(ctx context.Context, repo *Repository, from URL) {
	Pull(ctx, repo, from, mirrorRefSpecs)
}

func PushMirror(ctx context.Context, repo *Repository, to URL) {
	PushOnce(ctx, repo, to, mirrorRefSpecs)
}

func PullMirror(ctx context.Context, repo *Repository, from URL) {
	PullOnce(ctx, repo, from, mirrorRefSpecs)
}

func OverwriteRemote(ctx context.Context, repo *Repository, to URL, refspecs []config.RefSpec) (*git.Remote, string) {
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

func Push(ctx context.Context, repo *Repository, to URL, refspecs []config.RefSpec) {
	remote, remoteName := OverwriteRemote(ctx, repo, to, refspecs)
	must.NoError(ctx, remote.PushContext(ctx, &git.PushOptions{RemoteName: remoteName, Auth: GetAuth(ctx, to)}))
}

func Pull(ctx context.Context, repo *Repository, from URL, refspecs []config.RefSpec) {
	remote, remoteName := OverwriteRemote(ctx, repo, from, refspecs)
	must.NoError(ctx, remote.FetchContext(ctx, &git.FetchOptions{RemoteName: remoteName, Auth: GetAuth(ctx, from)}))
	err := remote.FetchContext(ctx, &git.FetchOptions{RemoteName: remoteName, Auth: GetAuth(ctx, from)})
	must.Assertf(ctx,
		err == transport.ErrEmptyRemoteRepository ||
			err == git.NoErrAlreadyUpToDate ||
			err == nil, "%v", err)
}

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
	must.NoError(ctx, remote.PushContext(ctx, &git.PushOptions{RemoteName: nonce, Auth: GetAuth(ctx, to)}))
}

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
	err := remote.FetchContext(ctx, &git.FetchOptions{RemoteName: nonce, Auth: GetAuth(ctx, from)})
	must.Assertf(ctx,
		err == transport.ErrEmptyRemoteRepository ||
			err == git.NoErrAlreadyUpToDate ||
			err == nil, "%v", err)
}
