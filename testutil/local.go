package testutil

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gov4git/lib4git/base"
	"github.com/gov4git/lib4git/git"
)

type PlainRepo struct {
	Dir  string
	Repo *git.Repository
}

func InitPlainRepo(t *testing.T, ctx context.Context) PlainRepo {
	repoDir := t.TempDir()
	base.Infof("repo %v", repoDir)
	repo := git.InitPlain(ctx, repoDir, false) // not bare
	return PlainRepo{Dir: repoDir, Repo: repo}
}

func Hang() {
	<-(chan int)(nil)
}

type LocalAddress struct {
	dir  string
	repo *git.Repository
	tree *git.Tree
	addr git.Address
}

func (x LocalAddress) String() string {
	return fmt.Sprintf("test address, dir=%v\n", x.dir)
}

func NewLocalAddress(ctx context.Context, t *testing.T, branch git.Branch, isBare bool) LocalAddress {
	dir := filepath.Join(t.TempDir(), UniqueString(ctx))
	repo := git.InitPlain(ctx, dir, isBare)
	addr := git.NewAddress(git.URL(dir), branch)
	var tree *git.Tree
	if !isBare {
		tree = git.Worktree(ctx, repo)
	}
	return LocalAddress{dir: dir, repo: repo, addr: addr, tree: tree}
}

func (x LocalAddress) Address() git.Address { return x.addr }

func (x LocalAddress) Push(context.Context) {}

func (x LocalAddress) Pull(context.Context) {}

func (x LocalAddress) Repo() *git.Repository { return x.repo }

func (x LocalAddress) Tree() *git.Tree { return x.tree }

func (x LocalAddress) Dir() string { return x.dir }
