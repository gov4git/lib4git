package git

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func TestMirror(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	dir1, dir2, dir3 := filepath.Join(dir, "1"), filepath.Join(dir, "2"), filepath.Join(dir, "3")
	fmt.Println("r1=", dir1)
	fmt.Println("r2=", dir2)
	fmt.Println("r3=", dir3)

	r1 := InitPlain(ctx, dir1, false)
	r2 := InitPlain(ctx, dir2, false)
	r3 := InitPlain(ctx, dir3, false)

	populate(ctx, r1, "ok1")
	populate(ctx, r2, "ok2")
	populate(ctx, r3, "ok3")

	Mirror(
		ctx,
		r1,
		[]string{"r2", "r3"},
		[]Address{
			{Repo: URL(dir2), Branch: Branch(MainBranch)},
			{Repo: URL(dir3), Branch: Branch(MainBranch)},
		},
		ns.NS("x/y/z"),
	)

	// TODO: add verification

	// <-(chan int)(nil)
}

func populate(ctx context.Context, r *git.Repository, nonce string) {
	w, _ := r.Worktree()
	f, _ := w.Filesystem.Create(nonce)
	f.Write([]byte(nonce))
	f.Close()
	_, err := w.Add(nonce)
	must.NoError(ctx, err)
	_, err = w.Commit(nonce, &git.CommitOptions{All: true})
	must.NoError(ctx, err)
}
