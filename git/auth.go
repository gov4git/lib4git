package git

import (
	"context"
	"sync"

	giturls "github.com/whilp/git-urls"

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

// repo-dependent auth

var authManager AuthManager

func GetAuth(ctx context.Context, forRepo URL) transport.AuthMethod {
	return authManager.GetAuth(ctx, forRepo)
}

type AuthManager struct {
	lk    sync.Mutex
	ssh   transport.AuthMethod
	https transport.AuthMethod
}

func (x *AuthManager) SetAuthHTTPS(a transport.AuthMethod) {
	x.lk.Lock()
	defer x.lk.Unlock()
	x.https = a
}

func (x *AuthManager) SetAuthSSH(a transport.AuthMethod) {
	x.lk.Lock()
	defer x.lk.Unlock()
	x.ssh = a
}

func (x *AuthManager) GetAuth(ctx context.Context, forRepo URL) transport.AuthMethod {
	u, err := giturls.Parse(string(forRepo))
	must.NoError(ctx, err)
	x.lk.Lock()
	defer x.lk.Unlock()
	switch u.Scheme {
	case "http", "https":
		return x.https
	case "ssh":
		return x.ssh
	default:
		return nil
	}
}
