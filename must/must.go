package must

import (
	"context"
	"fmt"
	"runtime/debug"
)

type Error struct {
	Ctx     context.Context
	Stack   []byte
	wrapped error
}

func (x *Error) Error() string {
	return x.Wrapped().Error()
}

func (x *Error) Wrapped() error {
	if x == nil {
		return nil
	}
	return x.wrapped
}

type ErrorString struct {
	Msg string `json:"msg"`
}

func (x ErrorString) Error() string {
	return x.Msg
}

func mkErr(ctx context.Context, wrapped error) *Error {
	return &Error{
		Ctx:     ctx,
		Stack:   debug.Stack(),
		wrapped: wrapped,
	}
}

func Panic(ctx context.Context, err error) {
	panic(mkErr(ctx, err))
}

func Errorf(ctx context.Context, format string, args ...any) {
	Panic(ctx, ErrorString{Msg: fmt.Sprintf(format, args...)})
}

func Assert(ctx context.Context, cond bool, err error) {
	if cond {
		return
	}
	Panic(ctx, err)
}

func Assertf(ctx context.Context, cond bool, format string, args ...any) {
	if cond {
		return
	}
	Errorf(ctx, format, args...)
}

func NoError(ctx context.Context, err error) {
	if err != nil {
		Panic(ctx, err)
	}
}

func Try(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(*Error).Wrapped()
		}
	}()
	f()
	return nil
}

func TryThru(f func()) (err *Error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(*Error)
		}
	}()
	f()
	return nil
}

func TryWithStack(f func()) (err error, stack []byte) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(*Error).Wrapped()
			stack = r.(*Error).Stack
		}
	}()
	f()
	return nil, nil
}

func Try1[R1 any](f func() R1) (_ R1, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(*Error).Wrapped()
		}
	}()
	return f(), nil
}

func Try1Thru[R1 any](f func() R1) (_ R1, err *Error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(*Error)
		}
	}()
	return f(), nil
}
