package git

import (
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

func init() {
	// client.InstallProtocol("file", server.NewClient(server.DefaultLoader))
	client.InstallProtocol("file", server.NewClient(server.NewFilesystemLoader(osfs.New(""))))
}
