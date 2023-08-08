package git

import "github.com/gov4git/lib4git/form"

type Change[Q any, R any] struct {
	Msg    string     `json:"msg"`
	Fn     string     `json:"fn"`
	Query  Q          `json:"query"`
	Result R          `json:"result"`
	Steps  form.Forms `json:"steps"`
}

type Commitable interface {
	Message() string
}

func (x Change[Q, R]) Message() string {
	return x.Msg
}

func NewChange[Q any, R any](
	msg string,
	fn string,
	query Q,
	result R,
	steps form.Forms,
) Change[Q, R] {
	return Change[Q, R]{
		Msg:    msg,
		Fn:     fn,
		Query:  query,
		Result: result,
		Steps:  steps,
	}
}

type ChangeNoResult = Change[form.None, form.None]

func NewChangeNoResult(msg string, fn string) ChangeNoResult {
	return NewChange(msg, fn, form.None{}, form.None{}, nil)
}
