package git

import "path/filepath"

func ListFilesRecursively(t *Tree, dir string) ([]string, error) {
	fs := t.Filesystem
	infos, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	list := []string{}
	for _, info := range infos {
		if info.IsDir() {
			sublist, err := ListFilesRecursively(t, filepath.Join(dir, info.Name()))
			if err != nil {
				return nil, err
			}
			list = append(list, sublist...)
		} else {
			list = append(list, filepath.Join(dir, info.Name()))
		}
	}
	return list, nil
}
