package git

import (
	"context"
	"math/rand"
	"path/filepath"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func Embed(
	ctx context.Context,
	repo *Repository,
	keys []string,
	addrs []Address,
	toBranch Branch,
	toNS []ns.NS,
	allowOverride bool,
) {

	// fetch remotes
	must.Assertf(ctx, len(toNS) == len(addrs) && len(keys) == len(addrs), "names and keys must match addresses")
	remoteTreeHashes := make([]plumbing.Hash, len(keys))
	remoteCommitHashes := make([]plumbing.Hash, len(keys))
	for i := range keys {
		remoteCommit := fetchEmbedding(ctx, repo, keys[i], addrs[i])
		remoteTreeHashes[i] = PrefixTree(ctx, repo, toNS[i], remoteCommit.TreeHash) // prefix with namespace
		remoteCommitHashes[i] = remoteCommit.Hash
	}

	// create a common tree with all embeddings merged together
	embeddingsTreeHash := MergeTrees(ctx, repo, remoteTreeHashes, allowOverride)

	// merge embeddings into the toBranch tree
	branchRefName := plumbing.NewBranchReferenceName(string(toBranch))
	branchRef := Reference(ctx, repo, branchRefName, true)
	branchCommit := GetCommit(ctx, repo, branchRef.Hash())
	mergedTreeHash := mergeTrees(ctx, repo, branchCommit.TreeHash, embeddingsTreeHash, allowOverride)

	// create a commit
	opts := git.CommitOptions{}
	must.NoError(ctx, opts.Validate(repo))
	commit := object.Commit{
		Author:       *opts.Author,
		Committer:    *opts.Committer,
		Message:      "embed remotes",
		TreeHash:     mergedTreeHash,
		ParentHashes: append([]plumbing.Hash{branchCommit.Hash}, remoteCommitHashes...),
	}
	commitObject := repo.Storer.NewEncodedObject()
	must.NoError(ctx, commit.Encode(commitObject))
	commitHash, err := repo.Storer.SetEncodedObject(commitObject)
	must.NoError(ctx, err)

	// update branch
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRefName, commitHash))
	must.NoError(ctx, err)
}

func fetchEmbedding(
	ctx context.Context,
	repo *Repository,
	key string,
	addr Address,
) object.Commit {

	// fetch remote branch using an ephemeral definition of the remote (not stored in the repo)
	nonce := "embedding-" + strconv.FormatUint(uint64(rand.Int63()), 36)
	remoteBranchName := plumbing.NewBranchReferenceName(string(addr.Branch))
	embeddedBranchName := plumbing.NewBranchReferenceName(filepath.Join("embedding", key))
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

func MergeTrees(
	ctx context.Context,
	repo *Repository,
	ths []plumbing.Hash,
	allowOverride bool,
) plumbing.Hash {

	aggregate := MakeTree(ctx, repo, object.Tree{})
	for _, th := range ths {
		aggregate = mergeTrees(ctx, repo, aggregate, th, allowOverride)
	}
	return aggregate
}

func mergeTrees(
	ctx context.Context,
	repo *Repository,
	leftTH plumbing.Hash, // TH = TreeHash
	rightTH plumbing.Hash,
	allowOverride bool,
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
		if left, ok := merged[right.Name]; ok {
			if left.Mode == filemode.Dir && right.Mode == filemode.Dir {
				// merge directories
				mergedLeftRightTH := mergeTrees(ctx, repo, left.Hash, right.Hash, allowOverride)
				merged[right.Name] = object.TreeEntry{Name: right.Name, Mode: filemode.Dir, Hash: mergedLeftRightTH}
			} else {
				// right overrides left
				must.Assertf(ctx, allowOverride, "not a mutually exclusive merge")
				merged[right.Name] = right
			}
		} else {
			merged[right.Name] = right
		}
	}

	// rebuild tree
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
