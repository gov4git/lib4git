package git

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/gov4git/lib4git/must"
)

var auth transport.AuthMethod

func setAuth(a transport.AuthMethod) {
	auth = a
}

func SetPasswordAuth(ctx context.Context, user string, pass string) {
	auth = &http.BasicAuth{Username: user, Password: pass}
}

func SetTokenAuth(ctx context.Context, token string) {
	auth = &http.BasicAuth{Username: "123", Password: token} // "123" can be anything but empty
}

func SetSSHFileAuth(ctx context.Context, user string, privKeyFile string) {
	pubKey, err := ssh.NewPublicKeysFromFile(user, privKeyFile, "")
	must.NoError(ctx, err)
	auth = pubKey
}
