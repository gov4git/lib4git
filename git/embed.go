package git

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func EmbedOnBranchReset(
	ctx context.Context,
	repo *Repository,
	addrs []Address, // remote branches to be embedded
	caches []Branch, // embedding cache branch name
	toBranch Branch, // embed into branch
	toNS []ns.NS, // namespace within into branch where each remote branch should be embedded
	allowOverride bool,
	filter MergeFilter,
) {

	EmbedOnBranch(ctx, repo, addrs, caches, toBranch, toNS, allowOverride, filter)
	ResetToBranch(ctx, repo, toBranch)
}

func EmbedOnBranch(
	ctx context.Context,
	repo *Repository,
	addrs []Address, // remote branches to be embedded
	caches []Branch, // embedding cache branch name
	toBranch Branch, // embed into branch
	toNS []ns.NS, // namespace within into branch where each remote branch should be embedded
	allowOverride bool,
	filter MergeFilter,
) plumbing.Hash {

	parentCommit := ResolveCreateBranch(ctx, repo, toBranch)
	h := EmbedOnCommit(ctx, repo, addrs, caches, parentCommit, toNS, allowOverride, filter)
	UpdateBranch(ctx, repo, toBranch, h)
	return h
}

// Embed creates a new commit on top of another one.
// The HEAD is not updated. The working tree is not updated.
func EmbedOnCommit(
	ctx context.Context,
	repo *Repository,
	addrs []Address, // remote branches to be embedded
	caches []Branch, // embedding cache branch name
	parentCommit *object.Commit,
	toNS []ns.NS, // namespace within into branch where each remote branch should be embedded
	allowOverride bool,
	filter MergeFilter,
) plumbing.Hash {

	// fetch remotes
	must.Assertf(ctx, len(toNS) == len(addrs), "namespaces and addresses must be same count")
	remoteTreeHashes := []plumbing.Hash{}
	remoteCommitHashes := []plumbing.Hash{}
	for i := range addrs {
		remoteCommit := fetchEmbedding(ctx, repo, addrs[i], caches[i])
		if remoteCommit == nil {
			fmt.Println("skipping empty repo", addrs[i])
			continue
		}
		fmt.Println("syncing", addrs[i])
		t := PrefixTree(ctx, repo, toNS[i], remoteCommit.TreeHash) // prefix with namespace
		remoteTreeHashes = append(remoteTreeHashes, t)
		remoteCommitHashes = append(remoteCommitHashes, remoteCommit.Hash)
	}

	// create a common tree with all embeddings merged together
	embeddingsTreeHash := MergeTrees(ctx, repo, remoteTreeHashes, allowOverride, filter)

	// merge embeddings into the toBranch tree
	// XXX: check if merge produced changes
	mergedTreeHash := mergeTrees(ctx, repo, ns.NS{}, parentCommit.TreeHash, embeddingsTreeHash, false, MergePassFilter)

	// create a commit
	parents := append([]plumbing.Hash{parentCommit.Hash}, remoteCommitHashes...)
	ch := CreateCommit(ctx, repo, "embed remotes", mergedTreeHash, parents)
	return ch
}

func fetchEmbedding(ctx context.Context, repo *Repository, addr Address, cache Branch) *object.Commit {

	// fetch remote branch using an ephemeral definition of the remote (not stored in the repo)
	nonce := "embedding-" + strconv.FormatUint(uint64(rand.Int63()), 36)
	remoteBranchName := plumbing.NewBranchReferenceName(string(addr.Branch))
	embeddedBranchName := plumbing.NewBranchReferenceName(string(cache))
	remote := git.NewRemote(
		repo.Storer,
		&config.RemoteConfig{
			Name: nonce,
			URLs: []string{string(addr.Repo)},
			Fetch: []config.RefSpec{
				config.RefSpec(remoteBranchName + ":" + embeddedBranchName),
			},
		},
	)
	err := remote.FetchContext(ctx, &git.FetchOptions{RemoteName: nonce})
	if err != nil {
		if !IsAlreadyUpToDate(err) && !IsRemoteRepoIsEmpty(err) {
			must.NoError(ctx, err)
		}
	}

	// get the latest commit on remote branch
	commitHash, err := repo.Reference(embeddedBranchName, true)
	if err == plumbing.ErrReferenceNotFound {
		return nil
	}
	must.NoError(ctx, err)
	commitObject := GetCommit(ctx, repo, commitHash.Hash())

	return commitObject
}
