package ns

import (
	"path/filepath"
	"strings"
)

// NS represents a namespace.
type NS []string

func ParseFromPath(p string) NS {
	return strings.Split(filepath.Clean(p), string(filepath.Separator))
}

func (ns NS) Path() string {
	return filepath.Join(ns...)
}

func (ns NS) Sub(path string) NS {
	sub := make(NS, len(ns)+1)
	for i := range ns {
		sub[i] = ns[i]
	}
	sub[len(ns)] = path
	return sub
}

func (ns NS) Join(sub NS) NS {
	x := NS{}
	x = append(x, ns...)
	x = append(x, sub...)
	return x
}

func (ns NS) Parts() []string {
	return ns
}
