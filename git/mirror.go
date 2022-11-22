package git

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func Mirror(ctx context.Context, repo *Repository, srcNames []string, srcAddrs []Address, rootNS ns.NS) {

	// fetch remotes
	must.Assertf(ctx, len(srcNames) == len(srcAddrs), "names must match addresses")
	remoteTreeHashes := make([]plumbing.Hash, len(srcNames))
	for i := range srcNames {
		remoteTreeHashes[i] = fetchMirror(ctx, repo, srcNames[i], srcAddrs[i], rootNS.Sub(srcNames[i]))
	}

	// create a common tree for all mirrors
	entries := make([]object.TreeEntry, len(srcNames))
	for i := range srcNames {
		entries[i] = object.TreeEntry{Name: srcNames[i], Mode: filemode.Dir, Hash: remoteTreeHashes[i]}
	}
	mirrorsTree := object.Tree{Entries: entries}
	treeObject := repo.Storer.NewEncodedObject()
	err := mirrorsTree.Encode(treeObject)
	must.NoError(ctx, err)
	mirrorsTreeHash, err := repo.Storer.SetEncodedObject(treeObject)
	must.NoError(ctx, err)

	// create a commit
	XXX
	commit := object.Commit{
		Author:       object.Signature{Name: "Petar", Email: "petar@example.com", When: time.Now()},
		Committer:    object.Signature{Name: "Petar", Email: "petar@example.com", When: time.Now()},
		Message:      "first commit",
		TreeHash:     rootHash,
		ParentHashes: []plumbing.Hash{c1.Hash, c2.Hash},
	}
	commitObject := repo.Storer.NewEncodedObject()
	err = commit.Encode(commitObject)
	must.NoError(ctx, err)
	commitHash, err := repo.Storer.SetEncodedObject(commitObject)
	must.NoError(ctx, err)

	// update HEAD
	updateHEAD(ctx, repo, commitHash)
}

func updateHEAD(ctx context.Context, repo *Repository, commitHash plumbing.Hash) {
	head, err := repo.Storer.Reference(plumbing.HEAD)
	must.NoError(ctx, err)

	name := plumbing.HEAD
	if head.Type() != plumbing.HashReference {
		name = head.Target()
	}

	ref := plumbing.NewHashReference(name, commitHash)
	err = repo.Storer.SetReference(ref)
	must.NoError(ctx, err)
}

func fetchMirror(ctx context.Context, repo *Repository, srcName string, srcAddr Address, dirNS ns.NS) plumbing.Hash {

	// fetch remote branch
	remoteName := "mirror-" + strconv.FormatUint(uint64(rand.Int63()), 36)
	remote := git.NewRemote(
		repo.Storer,
		&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{string(srcAddr.Repo)},
			Fetch: []config.RefSpec{
				config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/mirrors/%s", srcAddr.Branch, srcName)),
			},
		},
	)
	err := remote.FetchContext(ctx, &git.FetchOptions{RemoteName: remoteName})
	must.NoError(ctx, err)

	// get the hash of the latest commit on remote
	commitHash, err := repo.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/mirrors/%s", srcName)), true)
	must.NoError(ctx, err)

	// get tree from commits
	commitObject, err := object.GetCommit(repo.Storer, commitHash.Hash())
	must.NoError(ctx, err)

	return commitObject.TreeHash
}
