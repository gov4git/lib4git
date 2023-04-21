package git

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/gammazero/workerpool"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func EmbedOnBranchReset(
	ctx context.Context,
	repo *Repository,
	addrs []Address, // remote branches to be embedded
	caches []Branch, // embedding cache branch name
	toBranch Branch, // embed into branch
	toNS []ns.NS, // namespace within into branch where each remote branch should be embedded
	allowOverride bool,
	filter MergeFilter,
) {

	EmbedOnBranch(ctx, repo, addrs, caches, toBranch, toNS, allowOverride, filter)
	ResetToBranch(ctx, repo, toBranch)
}

func EmbedOnBranch(
	ctx context.Context,
	repo *Repository,
	addrs []Address, // remote branches to be embedded
	caches []Branch, // embedding cache branch name
	toBranch Branch, // embed into branch
	toNS []ns.NS, // namespace within into branch where each remote branch should be embedded
	allowOverride bool,
	filter MergeFilter,
) plumbing.Hash {

	parentCommit := ResolveCreateBranch(ctx, repo, toBranch)
	h := EmbedOnCommit(ctx, repo, addrs, caches, parentCommit, toNS, allowOverride, filter)
	if !h.IsZero() {
		UpdateBranch(ctx, repo, toBranch, h)
	}
	return h
}

// Embed creates a new commit on top of another one.
// The HEAD is not updated. The working tree is not updated.
func EmbedOnCommit(
	ctx context.Context,
	repo *Repository,
	addrs []Address, // remote branches to be embedded
	caches []Branch, // embedding cache branch name
	parentCommit *object.Commit,
	toNS []ns.NS, // namespace within into branch where each remote branch should be embedded
	allowOverride bool,
	filter MergeFilter,
) plumbing.Hash {

	// fetch remotes
	must.Assertf(ctx, len(toNS) == len(addrs), "namespaces and addresses must be same count")
	remoteTreeHashes := []plumbing.Hash{}
	remoteCommitHashes := []plumbing.Hash{}
	for i := range addrs {
		remoteCommit, err := FetchIntoBranch(ctx, repo, addrs[i], caches[i])
		if err != nil {
			fmt.Printf("skipping empty or inaccessible repo %v (%v)\n", addrs[i], err)
			continue
		}
		fmt.Println("syncing", addrs[i])
		t := PrefixTree(ctx, repo, toNS[i], remoteCommit.TreeHash) // prefix with namespace
		remoteTreeHashes = append(remoteTreeHashes, t)
		remoteCommitHashes = append(remoteCommitHashes, remoteCommit.Hash)
	}

	// create a common tree with all embeddings merged together
	embeddingsTreeHash := MergeTrees(ctx, repo, remoteTreeHashes, allowOverride, filter)

	// merge embeddings into the toBranch tree
	// XXX: check if merge produced changes, don't commit if it didn't
	mergedTreeHash := mergeTrees(ctx, repo, ns.NS{}, parentCommit.TreeHash, embeddingsTreeHash, false, MergePassFilter)

	// create a commit
	parents := append([]plumbing.Hash{parentCommit.Hash}, remoteCommitHashes...)
	ch := CreateCommit(ctx, repo, "embed remotes", mergedTreeHash, parents)
	return ch
}

// fetch all addresses to local disk proxy repos (parallel)
// fetch proxy repos into target repo (sequential)
/*
// repos must be different, because a Repository (a go-git Repository) is not thread-safe.
func FetchIntoBranches(ctx context.Context, repo []*Repository, addr []Address, toBranch []Branch) []error {

	// fetch addresses into local disk proxy repos
	localRepo := make([]*Repository, len(repo))
	localClone := make([]Cloned, len(repo))
	errs := make([]error, len(repo))
	// XXX: throttle
	for i := range repo {
		go func(i int) {
			localPath := filepath.Join(os.TempDir(), nonceName())
			localRepo[i] = openOrInitOnDisk(ctx, URL(localPath), true) // create a throw-away local disk repo
			_, err[i] = FetchIntoBranch(ctx, XXX)

			err := must.Try(
				func() {
					localClone[i] = CloneOneTo(ctx, addr[i], localRepo[i])
				},
			)
			if IsRepoIsInaccessible(err) {
				errs[i] = err
				return
			}
			if IsAlreadyUpToDate(err) || IsRemoteRepoIsEmpty(err) {
				errs[i] = err
				return
			}
			XXX
			// PullOnce(ctx, local[i], addr[i].Repo, clonePullRefSpecs(addr[i], false))
		}(i)
	}

	// commits := make([]*object.Commit, len(repo))
	// must.NoError(ctx, err)
	// commits[i], err[i] = LatestCommitOnBranch(ctx, XXX, XXX)
	// XXX: cleanup local repos
	return commits, errs
}
*/

func FetchIntoBranchPar(ctx context.Context, repo []*Repository, addr []Address, branch []Branch, maxPar int) ([]*object.Commit, []error) {
	commits := make([]*object.Commit, len(repo))
	errs := make([]error, len(repo))
	wp := workerpool.New(maxPar) // throttle
	for i_ := range repo {
		i := i_
		wp.Submit(
			func() {
				commits[i], errs[i] = FetchIntoBranch(ctx, repo[i], addr[i], branch[i])
			},
		)
	}
	wp.StopWait()
	return commits, errs
}

func FetchIntoBranch(ctx context.Context, repo *Repository, addr Address, branch Branch) (*object.Commit, error) {

	// fetch remote branch using an ephemeral definition of the remote (not stored in the repo)
	nonce := "embedding-" + strconv.FormatUint(uint64(rand.Int63()), 36)
	remoteBranchName := plumbing.NewBranchReferenceName(string(addr.Branch))
	embeddedBranchName := plumbing.NewBranchReferenceName(string(branch))
	remote := git.NewRemote(
		repo.Storer,
		&config.RemoteConfig{
			Name: nonce,
			URLs: []string{string(addr.Repo)},
			Fetch: []config.RefSpec{
				config.RefSpec(remoteBranchName + ":" + embeddedBranchName),
			},
		},
	)
	err := remote.FetchContext(ctx, &git.FetchOptions{
		RemoteName: nonce,
		Auth:       GetAuth(ctx, addr.Repo),
	})
	if IsRepoIsInaccessible(err) {
		return nil, err
	}
	if IsAlreadyUpToDate(err) || IsRemoteRepoIsEmpty(err) {
		return nil, err
	}
	must.NoError(ctx, err)

	return LatestCommitOnBranch(ctx, repo, Branch(embeddedBranchName))
}

// LatestCommitOnBranch returns the latest commit on a branch.
func LatestCommitOnBranch(ctx context.Context, repo *Repository, branch Branch) (*object.Commit, error) {
	commitHash, err := repo.Reference(branch.ReferenceName(), true)
	if IsRefNotFound(err) {
		return nil, err
	}
	must.NoError(ctx, err)
	commitObject := GetCommit(ctx, repo, commitHash.Hash())
	return commitObject, nil
}
