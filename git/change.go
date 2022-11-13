package git

import "github.com/gov4git/lib4git/form"

type Change[R any] struct {
	Result R
	Msg    string //commit msg
}

type ChangeNoResult = Change[form.None]
