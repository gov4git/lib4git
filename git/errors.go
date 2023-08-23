package git

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func IsAlreadyUpToDate(err error) bool {
	if err == nil {
		return false
	}
	return err == git.NoErrAlreadyUpToDate || err.Error() == "already up-to-date"
}

func IsRemoteRepoIsEmpty(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "remote repository is empty"
}

func IsAuthRequired(err error) bool {
	if err == nil {
		return false
	}
	return err == transport.ErrAuthenticationRequired
}

func IsIOTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), " i/o timeout")
}

func IsRepoNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err == transport.ErrRepositoryNotFound
}

func IsInvalidAuth(err error) bool {
	if err == nil {
		return false
	}
	return err == transport.ErrInvalidAuthMethod
}

func IsAuthFailed(err error) bool {
	if err == nil {
		return false
	}
	return err == transport.ErrAuthorizationFailed
}

func IsNoMatchingRefSpec(err error) bool {
	_, is := err.(git.NoMatchingRefSpecError)
	return is
}

func IsRefNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err == plumbing.ErrReferenceNotFound
}

func IsRepoIsInaccessible(err error) bool {
	return IsAuthRequired(err) || IsIOTimeout(err) || IsRepoNotFound(err)
}

func IsNotExist(err error) bool {
	return err == os.ErrNotExist
}

func IsNonFastForwardUpdate(err error) bool {
	return strings.HasPrefix(err.Error(), "non-fast-forward update")
}
