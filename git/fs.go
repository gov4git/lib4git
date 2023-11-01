package git

import (
	"github.com/gov4git/lib4git/ns"
)

func ListFilesRecursively(t *Tree, dir ns.NS) ([]ns.NS, error) {
	fs := t.Filesystem
	infos, err := fs.ReadDir(dir.GitPath())
	if err != nil {
		return nil, err
	}
	list := []ns.NS{}
	for _, info := range infos {
		childPath := dir.Append(info.Name())
		if info.IsDir() {
			sublist, err := ListFilesRecursively(t, childPath)
			if err != nil {
				return nil, err
			}
			list = append(list, sublist...)
		} else {
			list = append(list, childPath)
		}
	}
	return list, nil
}
