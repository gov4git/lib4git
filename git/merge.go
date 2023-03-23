package git

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/base"
	"github.com/gov4git/lib4git/ns"
)

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
	leftTree := GetTree(ctx, repo, leftTH)
	rightTree := GetTree(ctx, repo, rightTH)

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
