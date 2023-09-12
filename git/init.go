package git

import (
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

func init() {
	// fmt.Println("XXX")
	client.InstallProtocol("file", server.NewClient(server.DefaultLoader))
}
