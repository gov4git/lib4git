package git

import (
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	authorLk sync.Mutex
	author   *object.Signature = &object.Signature{
		Name:  "4git",
		Email: "no-reply@gov4git.xyz",
	}
)

func SetAuthor(name string, email string) {
	authorLk.Lock()
	defer authorLk.Unlock()
	author = &object.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}
}

func GetAuthor() *object.Signature {
	authorLk.Lock()
	defer authorLk.Unlock()
	return author
}
