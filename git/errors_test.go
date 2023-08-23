package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func TestNonFastForwardUpdateError(t *testing.T) {

	// create "remote" repo
	ctx := context.Background()
	dir := t.TempDir()
	fmt.Println(dir)
	repo := InitPlain(ctx, dir, true)
	address := Address{Repo: URL(dir), Branch: MainBranch}
	ChangeDefaultBranch(ctx, repo, MainBranch)

	// clone#1 and modify
	c1 := CloneOne(ctx, address)
	StringToFileStage(ctx, c1.Tree(), ns.NS{"file1"}, "value1")
	Commit(ctx, c1.Tree(), "c1")

	// clone#2 and modify
	c2 := CloneOne(ctx, address)
	StringToFileStage(ctx, c2.Tree(), ns.NS{"file1"}, "value2")
	Commit(ctx, c2.Tree(), "c2")

	// push#1
	c1.Push(ctx)

	// push#2
	err := must.Try(func() { c2.Push(ctx) })

	if !IsNonFastForwardUpdate(err) {
		t.Errorf("expecting non-fast-forward update error, got %v", err.Error())
	}
}

/*
NOTE: go-git does not support rebasing yet.

func TestFastForwardUpdate(t *testing.T) {

	// create "remote" repo
	ctx := context.Background()
	dir := t.TempDir()
	fmt.Println(dir)
	repo := InitPlain(ctx, dir, true)
	address := Address{Repo: URL(dir), Branch: MainBranch}
	ChangeDefaultBranch(ctx, repo, MainBranch)

	// clone#1 and modify
	c1 := CloneOne(ctx, address)
	StringToFileStage(ctx, c1.Tree(), ns.NS{"file1"}, "value1")
	Commit(ctx, c1.Tree(), "c1")

	// clone#2 and modify
	c2 := CloneOne(ctx, address)
	StringToFileStage(ctx, c2.Tree(), ns.NS{"file2"}, "value2")
	Commit(ctx, c2.Tree(), "c2")

	// push#1
	c1.Push(ctx)

	// push#2
	c2.Push(ctx)
}
*/
