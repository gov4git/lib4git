package git

import (
	"context"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/gov4git/lib4git/must"
)

func MakePasswordAuth(ctx context.Context, user string, pass string) transport.AuthMethod {
	return &http.BasicAuth{Username: user, Password: pass}
}

func MakeTokenAuth(ctx context.Context, token string) transport.AuthMethod {
	return &http.BasicAuth{Username: "123", Password: token} // "123" can be anything but empty
}

func MakeSSHFileAuth(ctx context.Context, user string, privKeyFile string) transport.AuthMethod {
	pubKey, err := ssh.NewPublicKeysFromFile(user, privKeyFile, "")
	must.NoError(ctx, err)
	return pubKey
}

// global repo-dependent auth

var authManager AuthManager

func SetAuth(forRepo URL, a transport.AuthMethod) {
	authManager.SetAuth(forRepo, a)
}

func GetAuth(ctx context.Context, forRepo URL) transport.AuthMethod {
	return authManager.GetAuth(ctx, forRepo)
}

type AuthManager struct {
	lk  sync.Mutex
	url map[URL]transport.AuthMethod
}

func (x *AuthManager) SetAuth(forRepo URL, a transport.AuthMethod) {
	x.lk.Lock()
	defer x.lk.Unlock()
	x.url[forRepo] = a
}

func (x *AuthManager) GetAuth(ctx context.Context, forRepo URL) transport.AuthMethod {
	x.lk.Lock()
	defer x.lk.Unlock()
	return x.url[forRepo]
}
