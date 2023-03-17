package git

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func ResolveCreateBranch(ctx context.Context, repo *Repository, branch Branch) *object.Commit {
	branchRef, err := repo.Reference(branch.ReferenceName(), true)
	if err == plumbing.ErrReferenceNotFound {
		CreateEmptyBranch(ctx, repo, branch)
		branchRef = Reference(ctx, repo, branch.ReferenceName(), true)
	} else {
		must.NoError(ctx, err)
	}
	return GetCommit(ctx, repo, branchRef.Hash())
}

func CreateEmptyBranch(ctx context.Context, repo *Repository, branch Branch) {
	th := MakeTree(ctx, repo, object.Tree{})
	ch := CreateCommit(ctx, repo, "init empty branch", th, nil)
	UpdateBranch(ctx, repo, branch, ch)
}

func CreateCommit(
	ctx context.Context,
	repo *Repository,
	msg string,
	treeHash plumbing.Hash,
	parents []plumbing.Hash,
) plumbing.Hash {

	opts := git.CommitOptions{Author: GetAuthor()}
	must.NoError(ctx, opts.Validate(repo))
	commit := object.Commit{
		Author:       *opts.Author,
		Committer:    *opts.Committer,
		Message:      msg,
		TreeHash:     treeHash,
		ParentHashes: parents,
	}
	commitObject := repo.Storer.NewEncodedObject()
	must.NoError(ctx, commit.Encode(commitObject))
	commitHash, err := repo.Storer.SetEncodedObject(commitObject)
	must.NoError(ctx, err)
	return commitHash
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
