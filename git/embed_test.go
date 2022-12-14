package git

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gov4git/lib4git/base"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func TestEmbed(t *testing.T) {
	base.LogVerbosely()
	ctx := context.Background()
	dir := t.TempDir()
	dir1, dir2, dir3 := filepath.Join(dir, "1"), filepath.Join(dir, "2"), filepath.Join(dir, "3")
	fmt.Println("r1=", dir1)
	fmt.Println("r2=", dir2)
	fmt.Println("r3=", dir3)

	r1 := InitPlain(ctx, dir1, false)
	r2 := InitPlain(ctx, dir2, false)
	r3 := InitPlain(ctx, dir3, false)

	populate(ctx, r1, "ok1", true)
	populate(ctx, r2, "ok2", true)
	populate(ctx, r3, "ok3", true)

	Embed(
		ctx,
		r1,
		[]string{"r2", "r3"},
		[]Address{
			{Repo: URL(dir2), Branch: Branch(MainBranch)},
			{Repo: URL(dir3), Branch: Branch(MainBranch)},
		},
		MainBranch,
		[]ns.NS{{"x", "y", "z", "r2"}, {"x", "y", "z", "r3"}},
		true,
	)

	populate(ctx, r1, "ha1", false)
	populate(ctx, r2, "ha2", false)
	populate(ctx, r3, "ha3", false)

	Embed(
		ctx,
		r1,
		[]string{"r2", "r3"},
		[]Address{
			{Repo: URL(dir2), Branch: Branch(MainBranch)},
			{Repo: URL(dir3), Branch: Branch(MainBranch)},
		},
		MainBranch,
		[]ns.NS{{"x", "y", "z", "r2"}, {"x", "y", "z", "r3"}},
		true,
	)

	// TODO: add verification

	// <-(chan int)(nil)
}

func populate(ctx context.Context, r *git.Repository, nonce string, createBranch bool) {
	w, err := r.Worktree()
	must.NoError(ctx, err)

	// point head to MainBranch
	branch := plumbing.NewBranchReferenceName(string(MainBranch))
	h := plumbing.NewSymbolicReference(plumbing.HEAD, branch)
	err = r.Storer.SetReference(h)
	must.NoError(ctx, err)

	if !createBranch {
		fmt.Println(branch)
		branchRef, err := r.Reference(branch, true)
		must.NoError(ctx, err)
		err = w.Reset(&git.ResetOptions{Commit: branchRef.Hash(), Mode: git.HardReset})
		must.NoError(ctx, err)
	}

	// make a change
	f, err := w.Filesystem.Create(nonce)
	must.NoError(ctx, err)
	f.Write([]byte(nonce))
	err = f.Close()
	must.NoError(ctx, err)
	_, err = w.Add(nonce)
	must.NoError(ctx, err)
	_, err = w.Commit(nonce, &git.CommitOptions{Author: GetAuthor()})
	must.NoError(ctx, err)
}
