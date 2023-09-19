//go:build linux || darwin

package git

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/base"
	"github.com/gov4git/lib4git/must"
)

var (
	testBranch  = Branch("test")
	test2Branch = Branch("test2")
	test3Branch = Branch("test3")
)

func TestCache(t *testing.T) {
	base.LogVerbosely()
	ctx := WithTTL(WithAuth(context.Background(), nil), nil)

	dir := t.TempDir()
	fmt.Println("test root ", dir)
	originDir := filepath.Join(dir, "origin")
	cacheDir := filepath.Join(dir, "cache")
	originAddr := Address{Repo: URL(originDir), Branch: testBranch}

	// init origin and cache
	InitPlain(ctx, originDir, true)
	cache := NewCache(ctx, cacheDir)

	cloned1 := cache.CloneOne(ctx, originAddr)
	populateNonce(ctx, cloned1.Repo(), "ok1")
	cloned1.Push(ctx)

	cloned2 := cache.CloneOne(ctx, originAddr)
	findFile(ctx, cloned2.Repo(), "ok1")
	populateNonce(ctx, cloned2.Repo(), "ok2")
	cloned2.Tree().Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(string(test2Branch)),
		Create: true,
	})
	populateNonce(ctx, cloned2.Repo(), "ok3")
	cloned2.Push(ctx)

	cloned3 := cache.CloneOne(ctx, Address{Repo: URL(originDir), Branch: test3Branch})
	populateNonce(ctx, cloned3.Repo(), "ok5")
	cloned3.Push(ctx)

	cloned4 := cache.CloneOne(ctx, Address{Repo: URL(originDir), Branch: test2Branch})
	findFile(ctx, cloned4.Repo(), "ok3")
	populateNonce(ctx, cloned4.Repo(), "ok6")
	cloned4.Push(ctx)

	// <-(chan int)(nil)
}

func populateNonce(ctx context.Context, r *git.Repository, nonce string) {
	w, err := r.Worktree()
	must.NoError(ctx, err)

	f, err := w.Filesystem.Create(nonce)
	must.NoError(ctx, err)
	f.Write([]byte(nonce))
	err = f.Close()
	must.NoError(ctx, err)
	_, err = w.Add(nonce)
	must.NoError(ctx, err)
	_, err = w.Commit(nonce, &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test", When: time.Now()},
	})
	must.NoError(ctx, err)
}

func findFile(ctx context.Context, r *git.Repository, filepath string) {
	w, err := r.Worktree()
	must.NoError(ctx, err)

	_, err = w.Filesystem.Stat(filepath)
	must.NoError(ctx, err)
}
