package git

import (
	"context"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/format/index"
	"github.com/gov4git/lib4git/form"
	"github.com/gov4git/lib4git/must"
)

func TreeMkdirAll(ctx context.Context, t *Tree, path string) {
	err := t.Filesystem.MkdirAll(path, 0755)
	must.NoError(ctx, err)
}

func ToFile[V form.Form](ctx context.Context, t *Tree, filePath string, value V) {
	TreeMkdirAll(ctx, t, filepath.Dir(filePath))
	form.ToFile(ctx, t.Filesystem, filePath, value)
}

func ToFileStage[V form.Form](ctx context.Context, t *Tree, filePath string, value V) {
	ToFile(ctx, t, filePath, value)
	Add(ctx, t, filePath)
}

func FromFile[V form.Form](ctx context.Context, t *Tree, filePath string) V {
	return form.FromFile[V](ctx, t.Filesystem, filePath)
}

func TryFromFile[V form.Form](ctx context.Context, t *Tree, filePath string) (v V, err error) {
	err = must.Try(
		func() {
			v = FromFile[V](ctx, t, filePath)
		},
	)
	return
}

func RenameStage(ctx context.Context, t *Tree, oldPath, newPath string) {
	stat, err := t.Filesystem.Stat(oldPath)
	must.NoError(ctx, err)
	if stat.IsDir() {
		renameStageDir(ctx, t, oldPath, newPath)
	} else {
		renameStageFile(ctx, t, oldPath, newPath)
	}
}

func renameStageDir(ctx context.Context, t *Tree, oldPath, newPath string) {
	infos, err := t.Filesystem.ReadDir(oldPath)
	must.NoError(ctx, err)
	for _, info := range infos {
		RenameStage(ctx, t, filepath.Join(oldPath, info.Name()), filepath.Join(newPath, info.Name()))
	}
}

func renameStageFile(ctx context.Context, t *Tree, oldPath, newPath string) {
	must.NoError(ctx, t.Filesystem.Rename(oldPath, newPath))
	_, err := t.Remove(oldPath)
	if err != index.ErrEntryNotFound {
		must.NoError(ctx, err)
	}
	_, err = t.Add(newPath)
	must.NoError(ctx, err)
}
