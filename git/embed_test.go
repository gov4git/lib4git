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

const testEmbedBranch Branch = "brew"

func TestEmbed(t *testing.T) {
	base.LogVerbosely()
	ctx := WithTTL(WithAuth(context.Background(), nil), nil)
	dir := t.TempDir()

	dir1, dir2, dir3 := filepath.Join(dir, "1"), filepath.Join(dir, "2"), filepath.Join(dir, "3")
	fmt.Println("r1=", dir1)
	fmt.Println("r2=", dir2)
	fmt.Println("r3=", dir3)

	InitPlain(ctx, dir1, true) // non-bare disk repo
	InitPlain(ctx, dir2, true)
	InitPlain(ctx, dir3, true)

	embed := func() {
		r := InitInMemory(ctx)
		PullAll(ctx, r, URL(dir1))

		EmbedOnBranch(
			ctx,
			r,
			[]Address{
				{Repo: URL(dir2), Branch: Branch(testEmbedBranch)},
				{Repo: URL(dir3), Branch: Branch(testEmbedBranch)},
			},
			[]Branch{
				"cache2",
				"cache3",
			},
			testEmbedBranch,
			[]ns.NS{{"embedded", "r2"}, {"embedded", "r3"}},
			true,
			MergePassFilter,
		)

		PushAll(ctx, r, URL(dir1))
	}

	populateRemote(ctx, dir1, "ok1")
	populateRemote(ctx, dir2, "ok2")
	populateRemote(ctx, dir3, "ok3")

	embed()
	// <-(chan int)(nil)

	findFileRemote(ctx, dir1, "embedded/r2/ok2")
	findFileRemote(ctx, dir1, "embedded/r3/ok3")

	populateRemote(ctx, dir1, "ha1")
	populateRemote(ctx, dir2, "ha2")
	populateRemote(ctx, dir3, "ha3")
	embed()
	findFileRemote(ctx, dir1, "embedded/r2/ha2")
	findFileRemote(ctx, dir1, "embedded/r3/ha3")

	// <-(chan int)(nil)
}

func populate(ctx context.Context, r *git.Repository, nonce string) {
	w, err := r.Worktree()
	must.NoError(ctx, err)

	// point head to testEmbedBranch
	branch := plumbing.NewBranchReferenceName(string(testEmbedBranch))
	h := plumbing.NewSymbolicReference(plumbing.HEAD, branch)
	err = r.Storer.SetReference(h)
	must.NoError(ctx, err)

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

func populateRemote(ctx context.Context, p string, nonce string) {
	r := InitInMemory(ctx)
	PullAll(ctx, r, URL(p))
	populate(ctx, r, nonce)
	PushAll(ctx, r, URL(p))
}

func findFileRemote(ctx context.Context, p string, filepath string) {
	r := InitInMemory(ctx)
	PullAll(ctx, r, URL(p))
	w := Worktree(ctx, r)
	err := w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(string(testEmbedBranch))})
	must.NoError(ctx, err)

	findFile(ctx, r, filepath)
}

func findFile(ctx context.Context, r *git.Repository, filepath string) {
	w, err := r.Worktree()
	must.NoError(ctx, err)

	_, err = w.Filesystem.Stat(filepath)
	must.NoError(ctx, err)
}
