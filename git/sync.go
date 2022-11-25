package git

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/gov4git/lib4git/must"
)

func nonceName() string {
	return "nonce-" + strconv.FormatUint(uint64(rand.Int63()), 36)
}

func MirrorRefSpecXXX(branch Branch) config.RefSpec {
	return config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/mirrors/%s", branch, branch))
}

func PushRepoToURL(ctx context.Context, from *Repository, to URL, refspecs []config.RefSpec) {
	nonce := nonceName()
	remote := git.NewRemote(
		from.Storer,
		&config.RemoteConfig{
			Name:  nonce,
			URLs:  []string{string(to)},
			Fetch: refspecs,
		},
	)
	must.NoError(ctx, remote.PushContext(ctx, &git.PushOptions{RemoteName: nonce}))
}
