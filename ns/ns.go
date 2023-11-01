package ns

import (
	"path/filepath"
	"slices"
	"strings"
)

// NS represents a namespace.
type NS []string

func Equal(a, b NS) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ParseFromOSPath(p string) NS {
	return strings.Split(filepath.ToSlash(filepath.Clean(p)), "/")
}

func ParseFromGitPath(p string) NS {
	return strings.Split(p, "/")
}

func (ns NS) GitPath() string {
	return strings.Join(ns, "/")
}

func (ns NS) OSPath() string {
	return filepath.Join(ns...)
}

func (ns NS) Ext(ext string) NS {
	if len(ns) == 0 {
		return NS{"." + ext}
	}
	xs := make(NS, len(ns))
	copy(xs, ns)
	xs[len(xs)-1] = xs[len(xs)-1] + "." + ext
	return xs
}

// Deprecated: Use Append.
func (ns NS) Sub(pathElem string) NS {
	return ns.Append(pathElem)
}

func (ns NS) Append(elems ...string) NS {
	r := slices.Clone(ns)
	return append(r, elems...)
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

func (ns NS) Base() string {
	if len(ns) == 0 {
		return ""
	}
	return ns[len(ns)-1]
}

func (ns NS) Dir() NS {
	if len(ns) == 0 {
		return NS{}
	}
	return ns[:len(ns)-1]
}

func (ns NS) Len() int {
	return len(ns)
}
