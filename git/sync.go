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

var mirrorRefSpecs = []config.RefSpec{mirrorBranchesRefSpec, mirrorTagsRefSpec}

func PushMirror(ctx context.Context, repo *Repository, to URL) {
	PushRefSpecs(ctx, repo, to, mirrorRefSpecs)
}

func PullMirror(ctx context.Context, repo *Repository, from URL) {
	PullRefSpecs(ctx, repo, from, mirrorRefSpecs)
}

func PushRefSpecs(ctx context.Context, repo *Repository, to URL, refspecs []config.RefSpec) {
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

func PullRefSpecs(ctx context.Context, repo *Repository, from URL, refspecs []config.RefSpec) {
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
