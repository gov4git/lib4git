package file

import (
	"context"
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func StringToFile(ctx context.Context, fs billy.Filesystem, path ns.NS, content string) {
	BytesToFile(ctx, fs, path, []byte(content))
}

func FileToString(ctx context.Context, fs billy.Filesystem, path ns.NS) string {
	return string(FileToBytes(ctx, fs, path))
}

func BytesToFile(ctx context.Context, fs billy.Filesystem, path ns.NS, content []byte) {
	file, err := fs.OpenFile(path.GitPath(), os.O_CREATE|os.O_RDWR, 0644)
	must.NoError(ctx, err)
	defer file.Close()
	_, err = file.Write(content)
	must.NoError(ctx, err)
}

func FileToBytes(ctx context.Context, fs billy.Filesystem, path ns.NS) []byte {
	file, err := fs.Open(path.GitPath())
	must.NoError(ctx, err)
	content, err := io.ReadAll(file)
	must.NoError(ctx, err)
	return content
}
