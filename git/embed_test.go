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

	r1 := InitPlain(ctx, dir1, false) // non-bare disk repo
	r2 := InitPlain(ctx, dir2, false)
	r3 := InitPlain(ctx, dir3, false)

	embed := func() {
		EmbedOnBranchReset(
			ctx,
			r1,
			[]Address{
				{Repo: URL(dir2), Branch: Branch(MainBranch)},
				{Repo: URL(dir3), Branch: Branch(MainBranch)},
			},
			[]Branch{
				"cache2",
				"cache3",
			},
			MainBranch,
			[]ns.NS{{"embedded", "r2"}, {"embedded", "r3"}},
			true,
			MergePassFilter,
		)
	}

	populate(ctx, r1, "ok1")
	populate(ctx, r2, "ok2")
	populate(ctx, r3, "ok3")
	embed()
	findFile(ctx, r1, "embedded/r2/ok2")
	findFile(ctx, r1, "embedded/r3/ok3")

	populate(ctx, r1, "ha1")
	populate(ctx, r2, "ha2")
	populate(ctx, r3, "ha3")
	embed()
	findFile(ctx, r1, "embedded/r2/ha2")
	findFile(ctx, r1, "embedded/r3/ha3")

	// <-(chan int)(nil)
}

func populate(ctx context.Context, r *git.Repository, nonce string) {
	w, err := r.Worktree()
	must.NoError(ctx, err)

	// point head to MainBranch
	branch := plumbing.NewBranchReferenceName(string(MainBranch))
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
