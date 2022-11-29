package git

import (
	"context"
	"fmt"
	"testing"

	"github.com/gov4git/lib4git/must"
	giturls "github.com/whilp/git-urls"
)

func TestAuthURL(t *testing.T) {
	ctx := context.Background()

	u, err := giturls.Parse("git@github.com:petar/gov4git.public.git")
	must.NoError(ctx, err)
	fmt.Println(u)

	u, err = giturls.Parse("/x/y/z")
	must.NoError(ctx, err)
	fmt.Println(u)
}
