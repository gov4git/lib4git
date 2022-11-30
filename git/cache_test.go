package git

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gov4git/lib4git/base"
)

func TestCache(t *testing.T) {
	base.LogVerbosely()
	ctx := context.Background()

	dir := t.TempDir()
	fmt.Println("test root ", dir)
	originDir := filepath.Join(dir, "origin")
	cacheDir := filepath.Join(dir, "cache")
	originAddr := Address{Repo: URL(originDir), Branch: MainBranch}

	InitPlain(ctx, originDir, true)
	cache := NewCache(ctx, cacheDir)
	cloned1 := cache.Clone(ctx, originAddr)
	populate(ctx, cloned1.Repo(), "hi", true)
	cloned1.Push(ctx)

	cloned2 := cache.Clone(ctx, originAddr)
	populate(ctx, cloned2.Repo(), "ok", false)
	cloned2.Push(ctx)

	// <-(chan int)(nil)
}
