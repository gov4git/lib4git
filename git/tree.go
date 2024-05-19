package git

import (
	"context"
	"io/fs"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/gov4git/lib4git/file"
	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/must"
	"github.com/gov4git/lib4git/ns"
)

func TreeMkdirAll(ctx context.Context, t *Tree, path ns.NS) {
	if path.Len() == 0 {
		return
	}
	err := t.Filesystem.MkdirAll(path.GitPath(), 0755)
	must.NoError(ctx, err)
}

func TreeStat(ctx context.Context, t *Tree, path ns.NS) (fs.FileInfo, error) {
	return t.Filesystem.Stat(path.GitPath())
}

func TreeRemove(ctx context.Context, t *Tree, path ns.NS) (plumbing.Hash, error) {
	return t.Remove(path.GitPath())
}

func TreeReadDir(ctx context.Context, t *Tree, path ns.NS) ([]fs.FileInfo, error) {
	return t.Filesystem.ReadDir(path.GitPath())
}

func TreeStageAll(ctx context.Context, t *Tree) {
	addOpts := &gogit.AddOptions{All: true, Path: "/"}
	must.NoError(ctx, t.AddWithOptions(addOpts)) // XXX: check it works
}

//

func BytesToFile(ctx context.Context, t *Tree, path ns.NS, content []byte) {
	TreeMkdirAll(ctx, t, path.Dir())
	file.BytesToFile(ctx, t.Filesystem, path, content)
}

func BytesToFileStage(ctx context.Context, t *Tree, path ns.NS, content []byte) {
	BytesToFile(ctx, t, path, content)
	Add(ctx, t, path)
}

func FileToBytes(ctx context.Context, t *Tree, path ns.NS) []byte {
	return file.FileToBytes(ctx, t.Filesystem, path)
}

//

func StringToFile(ctx context.Context, t *Tree, path ns.NS, content string) {
	TreeMkdirAll(ctx, t, path.Dir())
	file.StringToFile(ctx, t.Filesystem, path, content)
}

func StringToFileStage(ctx context.Context, t *Tree, path ns.NS, content string) {
	StringToFile(ctx, t, path, content)
	Add(ctx, t, path)
}

func FileToString(ctx context.Context, t *Tree, path ns.NS) string {
	return file.FileToString(ctx, t.Filesystem, path)
}

//

func ToFile[V form.Form](ctx context.Context, t *Tree, filePath ns.NS, value V) {
	TreeMkdirAll(ctx, t, filePath.Dir())
	form.ToFile(ctx, t.Filesystem, filePath, value)
}

func ToFileStage[V form.Form](ctx context.Context, t *Tree, filePath ns.NS, value V) {
	ToFile(ctx, t, filePath, value)
	Add(ctx, t, filePath)
}

func FromFile[V form.Form](ctx context.Context, t *Tree, filePath ns.NS) V {
	return form.FromFile[V](ctx, t.Filesystem, filePath)
}

func FromFileInto(ctx context.Context, t *Tree, filePath ns.NS, into form.Form) {
	form.FromFileInto(ctx, t.Filesystem, filePath, into)
}

func TryFromFile[V form.Form](ctx context.Context, t *Tree, filePath ns.NS) (v V, err error) {
	err = must.Try(
		func() {
			v = FromFile[V](ctx, t, filePath)
		},
	)
	return
}

func RenameStage(ctx context.Context, t *Tree, oldPath, newPath ns.NS) {
	stat, err := t.Filesystem.Stat(oldPath.GitPath())
	must.NoError(ctx, err)
	if stat.IsDir() {
		renameStageDir(ctx, t, oldPath, newPath)
	} else {
		renameStageFile(ctx, t, oldPath, newPath)
	}
}

func renameStageDir(ctx context.Context, t *Tree, oldPath, newPath ns.NS) {
	infos, err := t.Filesystem.ReadDir(oldPath.GitPath())
	must.NoError(ctx, err)
	for _, info := range infos {
		RenameStage(ctx, t, oldPath.Sub(info.Name()), newPath.Sub(info.Name()))
	}
}

func renameStageFile(ctx context.Context, t *Tree, oldPath, newPath ns.NS) {
	must.NoError(ctx, t.Filesystem.Rename(oldPath.GitPath(), newPath.GitPath()))
	_, err := t.Remove(oldPath.GitPath())
	if err != index.ErrEntryNotFound {
		must.NoError(ctx, err)
	}
	_, err = t.Add(newPath.GitPath())
	must.NoError(ctx, err)
}
