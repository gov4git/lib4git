package ns

import (
	"path/filepath"
	"strings"
)

type NS string

func (ns NS) Path() string {
	return filepath.Clean(string(ns))
}

func (ns NS) Sub(path string) NS {
	return NS(filepath.Join(string(ns), path))
}

func (ns NS) Join(sub NS) NS {
	return ns.Sub(string(sub))
}

func (ns NS) Parts() []string {
	return strings.Split(ns.Path(), "/") //XXX: ns must be OS independent
}
