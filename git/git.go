package git

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

type URL string

func (u URL) Hash() string {
	return form.StringHashForFilename(string(u))
}

type Branch string

func (b Branch) ReferenceName() plumbing.ReferenceName {
	return plumbing.ReferenceName(plumbing.NewBranchReferenceName(string(b)))
}

func (b Branch) Sub(s string) Branch {
	return Branch(filepath.Join(string(b), s))
}

type CommitHash string

const MainBranch Branch = "main"

const Origin = "origin"

type Address struct {
	Repo   URL
	Branch Branch
}

func (a Address) Sub(s string) Address {
	return Address{Repo: a.Repo, Branch: a.Branch.Sub(s)}
}

func (a Address) Join(s ns.NS) Address {
	return Address{Repo: a.Repo, Branch: a.Branch.Sub(s.Path())}
}

func (a Address) String() string {
	return string(a.Repo) + ":" + string(a.Branch)
}

func (a Address) Hash() string {
	return form.StringHashForFilename(a.String())
}

func NewAddress(repo URL, branch Branch) Address {
	return Address{Repo: repo, Branch: branch}
}

type Repository = git.Repository

type Tree = git.Worktree

type RepoTree struct {
	Repo *Repository
	Tree *Tree
}

func SetHeadToBranch(ctx context.Context, repo *Repository, branch Branch) {
	branchName := plumbing.NewBranchReferenceName(string(branch))
	h := plumbing.NewSymbolicReference(plumbing.HEAD, branchName)
	err := repo.Storer.SetReference(h)
	must.NoError(ctx, err)
}

func ChangeDefaultBranch(ctx context.Context, repo *Repository, default_ Branch) {
	// https://github.com/hairyhenderson/gomplate/pull/1217/files#diff-06d907e05a1688ce7548c3d8b4877a01a61b3db506755db4419761dbe9fe0a5bR232
	SetHeadToBranch(ctx, repo, default_)

	c, err := repo.Config()
	must.NoError(ctx, err)
	c.Init.DefaultBranch = string(default_)

	err = repo.Storer.SetConfig(c)
	must.NoError(ctx, err)
}

func InitPlain(ctx context.Context, path string, isBare bool) *Repository {
	repo, err := git.PlainInit(path, isBare)
	if err != nil {
		must.Panic(ctx, err)
	}
	ChangeDefaultBranch(ctx, repo, MainBranch)
	return repo
}

func Worktree(ctx context.Context, repo *Repository) *Tree {
	wt, err := repo.Worktree()
	if err != nil {
		must.Panic(ctx, err)
	}
	return wt
}

func Add(ctx context.Context, wt *Tree, path string) {
	if _, err := wt.Add(path); err != nil {
		must.Panic(ctx, err)
	}
}

func Commit(ctx context.Context, wt *Tree, msg string) {
	if _, err := wt.Commit(msg, &git.CommitOptions{Author: GetAuthor()}); err != nil {
		must.Panic(ctx, err)
	}
}

func Checkout(ctx context.Context, wt *Tree, branch Branch) {
	err := wt.Checkout(
		&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(string(branch)),
		},
	)
	must.NoError(ctx, err)
}

func Reference(ctx context.Context, r *Repository, name plumbing.ReferenceName, resolved bool) *plumbing.Reference {
	x, err := r.Reference(name, resolved)
	must.NoError(ctx, err)
	return x
}

func Head(ctx context.Context, r *Repository) CommitHash {
	h, err := r.Head()
	must.NoError(ctx, err)
	return CommitHash(h.Hash().String())
}

func Remotes(ctx context.Context, r *Repository) []*git.Remote {
	h, err := r.Remotes()
	must.NoError(ctx, err)
	return h
}

func Branches(ctx context.Context, r *Repository) []*plumbing.Reference {
	iter, err := r.Branches()
	must.NoError(ctx, err)
	refs := []*plumbing.Reference{}
	for {
		ref, err := iter.Next()
		if err != nil {
			break
		}
		refs = append(refs, ref)
	}
	return refs
}

func Dump(ctx context.Context, r *Repository) {
	iter, err := r.Log(&git.LogOptions{})
	if err != nil {
		fmt.Println("HEAD not found")
		return
	}
	fmt.Println("Log:")
	for {
		c, err := iter.Next()
		if err != nil {
			break
		}
		fmt.Println(c)
	}
	fmt.Println("HEAD:", Head(ctx, r))
	for _, r := range Remotes(ctx, r) {
		fmt.Println("REMOTE:", r)
	}
	for _, r := range Branches(ctx, r) {
		fmt.Println("BRANCH:", r)
	}
}

func ResolveBranch(ctx context.Context, repo *Repository, branch Branch) *object.Commit {
	branchRef := Reference(ctx, repo, branch.ReferenceName(), true)
	return GetCommit(ctx, repo, branchRef.Hash())
}

func UpdateBranch(ctx context.Context, repo *Repository, branch Branch, h plumbing.Hash) {
	err := repo.Storer.SetReference(plumbing.NewHashReference(branch.ReferenceName(), h))
	must.NoError(ctx, err)
}

func ResetToBranch(ctx context.Context, repo *Repository, branch Branch) {
	w := Worktree(ctx, repo)
	branchRef := Reference(ctx, repo, branch.ReferenceName(), true)
	err := w.Reset(&git.ResetOptions{Commit: branchRef.Hash(), Mode: git.HardReset})
	must.NoError(ctx, err)
}
