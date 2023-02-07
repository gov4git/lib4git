package git

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/base"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func EmbedReset(
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

	w, err := repo.Worktree()
	must.NoError(ctx, err)
	branchRef := Reference(ctx, repo, toBranch.ReferenceName(), true)
	err = w.Reset(&git.ResetOptions{Commit: branchRef.Hash(), Mode: git.HardReset})
	must.NoError(ctx, err)
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

	branchRef := Reference(ctx, repo, toBranch.ReferenceName(), true)
	parentCommit := GetCommit(ctx, repo, branchRef.Hash())
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
	remoteTreeHashes := make([]plumbing.Hash, len(addrs))
	remoteCommitHashes := make([]plumbing.Hash, len(addrs))
	for i := range addrs {
		remoteCommit := fetchEmbedding(ctx, repo, addrs[i], caches[i])
		remoteTreeHashes[i] = PrefixTree(ctx, repo, toNS[i], remoteCommit.TreeHash) // prefix with namespace
		remoteCommitHashes[i] = remoteCommit.Hash
	}

	// create a common tree with all embeddings merged together
	embeddingsTreeHash := MergeTrees(ctx, repo, remoteTreeHashes, allowOverride, filter)

	// merge embeddings into the toBranch tree
	mergedTreeHash := mergeTrees(ctx, repo, ns.NS{}, parentCommit.TreeHash, embeddingsTreeHash, false, MergePassFilter)

	// create a commit
	opts := git.CommitOptions{Author: GetAuthor()}
	must.NoError(ctx, opts.Validate(repo))
	commit := object.Commit{
		Author:       *opts.Author,
		Committer:    *opts.Committer,
		Message:      "embed remotes",
		TreeHash:     mergedTreeHash,
		ParentHashes: append([]plumbing.Hash{parentCommit.Hash}, remoteCommitHashes...),
	}
	commitObject := repo.Storer.NewEncodedObject()
	must.NoError(ctx, commit.Encode(commitObject))
	commitHash, err := repo.Storer.SetEncodedObject(commitObject)
	must.NoError(ctx, err)

	return commitHash
}

func fetchEmbedding(ctx context.Context, repo *Repository, addr Address, cache Branch) object.Commit {

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
	must.NoError(ctx, remote.FetchContext(ctx, &git.FetchOptions{RemoteName: nonce}))

	// get the latest commit on remote branch
	commitHash := Reference(ctx, repo, embeddedBranchName, true)
	commitObject := GetCommit(ctx, repo, commitHash.Hash())

	return *commitObject
}

type MergeFilter func(ns.NS, object.TreeEntry) bool

func MergePassFilter(fromNS ns.NS, fromEntry object.TreeEntry) bool {
	return true
}

func MergeTrees(
	ctx context.Context,
	repo *Repository,
	ths []plumbing.Hash,
	allowOverride bool,
	filter MergeFilter,
) plumbing.Hash {

	aggregate := MakeTree(ctx, repo, object.Tree{})
	for _, th := range ths {
		aggregate = mergeTrees(ctx, repo, ns.NS{}, aggregate, th, allowOverride, filter)
	}
	return aggregate
}

func mergeTrees(
	ctx context.Context,
	repo *Repository,
	ns ns.NS,
	leftTH plumbing.Hash, // TH = TreeHash
	rightTH plumbing.Hash,
	allowOverride bool,
	rightFilter MergeFilter,
) plumbing.Hash {

	// get trees
	leftTree, err := object.GetTree(repo.Storer, leftTH)
	must.NoError(ctx, err)
	rightTree, err := object.GetTree(repo.Storer, rightTH)
	must.NoError(ctx, err)

	// merge tree entries
	merged := map[string]object.TreeEntry{}
	for _, left := range leftTree.Entries {
		merged[left.Name] = left
	}
	for _, right := range rightTree.Entries {
		if !rightFilter(ns, right) {
			continue
		}
		if left, ok := merged[right.Name]; ok {
			if left.Mode == filemode.Dir && right.Mode == filemode.Dir {
				// merge directories
				mergedLeftRightTH := mergeTrees(ctx, repo, ns.Sub(right.Name), left.Hash, right.Hash, allowOverride, rightFilter)
				merged[right.Name] = object.TreeEntry{Name: right.Name, Mode: filemode.Dir, Hash: mergedLeftRightTH}
			} else {
				// right overrides left
				if allowOverride {
					merged[right.Name] = right
				} else {
					base.Infof("tree entry %v already exists", ns.Sub(right.Name))
				}
			}
		} else {
			merged[right.Name] = right
		}
	}

	// make tree
	entries := make([]object.TreeEntry, 0, len(merged))
	for _, mergedEntry := range merged {
		entries = append(entries, mergedEntry)
	}
	return MakeTree(ctx, repo, object.Tree{Entries: entries})
}

func MakeTree(ctx context.Context, repo *Repository, tree object.Tree) plumbing.Hash {
	treeObject := repo.Storer.NewEncodedObject()
	err := tree.Encode(treeObject)
	must.NoError(ctx, err)
	treeHash, err := repo.Storer.SetEncodedObject(treeObject)
	must.NoError(ctx, err)
	return treeHash
}

// PrefixTree creates a git tree containing the tree th at path prefix.
func PrefixTree(
	ctx context.Context,
	repo *Repository,
	prefix ns.NS,
	th plumbing.Hash,
) plumbing.Hash {

	if len(prefix) == 0 {
		return th
	}

	mergedTree := object.Tree{
		Entries: []object.TreeEntry{
			{
				Name: prefix[0],
				Mode: filemode.Dir,
				Hash: PrefixTree(ctx, repo, prefix[1:], th),
			},
		},
	}
	treeObject := repo.Storer.NewEncodedObject()
	err := mergedTree.Encode(treeObject)
	must.NoError(ctx, err)
	prefixedTreeHash, err := repo.Storer.SetEncodedObject(treeObject)
	must.NoError(ctx, err)

	return prefixedTreeHash
}
