package git

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/must"
)

func GetCommit(ctx context.Context, r *Repository, h plumbing.Hash) *object.Commit {
	c, err := object.GetCommit(r.Storer, h)
	must.NoError(ctx, err)
	return c
}
