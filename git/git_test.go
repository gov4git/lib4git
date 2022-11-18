package git

import (
	"context"
	"fmt"
	"testing"
)

func TestRenameMain(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	fmt.Println(dir)
	repo := InitPlain(ctx, dir, false)
	ChangeDefaultBranch(ctx, repo, MainBranch)
}

// TODO: test fetching from a private git
