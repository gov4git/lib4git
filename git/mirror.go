package git

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func Mirror(
	ctx context.Context,
	repo *Repository,
	srcNames []string,
	srcAddrs []Address,
	toBranch Branch,
	toNS ns.NS,
) {

	// fetch remotes
	must.Assertf(ctx, len(srcNames) == len(srcAddrs), "names must match addresses")
	remoteTreeHashes := make([]plumbing.Hash, len(srcNames))
	remoteCommitHashes := make([]plumbing.Hash, len(srcNames))
	for i := range srcNames {
		remoteCommit := fetchMirror(ctx, repo, srcNames[i], srcAddrs[i])
		remoteTreeHashes[i] = remoteCommit.TreeHash
		remoteCommitHashes[i] = remoteCommit.Hash
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

	// merge mirrors into the toBranch tree
	branchRefName := plumbing.NewBranchReferenceName(string(toBranch))
	branchRef, err := repo.Reference(branchRefName, true)
	must.NoError(ctx, err)
	branchCommitObject, err := object.GetCommit(repo.Storer, branchRef.Hash())
	must.NoError(ctx, err)
	mergedTreeHash := attachTreeAtPath(ctx, repo, branchCommitObject.TreeHash, toNS, mirrorsTreeHash)

	// create a commit
	opts := git.CommitOptions{}
	err = opts.Validate(repo)
	must.NoError(ctx, err)
	commit := object.Commit{
		Author:       *opts.Author,
		Committer:    *opts.Committer,
		Message:      "merge mirrors",
		TreeHash:     mergedTreeHash,
		ParentHashes: append([]plumbing.Hash{branchCommitObject.Hash}, remoteCommitHashes...),
	}
	commitObject := repo.Storer.NewEncodedObject()
	err = commit.Encode(commitObject)
	must.NoError(ctx, err)
	commitHash, err := repo.Storer.SetEncodedObject(commitObject)
	must.NoError(ctx, err)

	// update branch
	err = repo.Storer.SetReference(plumbing.NewHashReference(branchRefName, commitHash))
	must.NoError(ctx, err)
}

func fetchMirror(
	ctx context.Context,
	repo *Repository,
	srcName string,
	srcAddr Address,
) object.Commit {

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

	return *commitObject
}

func mergeTrees(
	ctx context.Context,
	repo *Repository,
	leftTH plumbing.Hash, // TH = TreeHash
	rightTH plumbing.Hash,
	rightOverrides bool,
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
				mergedLeftRightTH := mergeTrees(ctx, repo, left.Hash, right.Hash, rightOverrides)
				merged[right.Name] = object.TreeEntry{Name: right.Name, Mode: filemode.Dir, Hash: mergedLeftRightTH}
			} else {
				// right overrides left
				must.Assertf(ctx, rightOverrides, "not a mutually exclusive merge")
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
	mergedTree := object.Tree{Entries: entries}
	treeObject := repo.Storer.NewEncodedObject()
	err = mergedTree.Encode(treeObject)
	must.NoError(ctx, err)
	mergedTreeHash, err := repo.Storer.SetEncodedObject(treeObject)
	must.NoError(ctx, err)

	return mergedTreeHash
}

func prefixTree(
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
				Mode: filemode.Dir, // XXX: always a dir?
				Hash: prefixTree(ctx, repo, prefix[1:], th),
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

func attachTreeAtPath(
	ctx context.Context,
	repo *Repository,
	rootTH plumbing.Hash, // TH = TreeHash
	atPath []string,
	attachTH plumbing.Hash,
) (newRootTH plumbing.Hash) {

	return mergeTrees(ctx, repo, rootTH, prefixTree(ctx, repo, atPath, attachTH), true)
}
