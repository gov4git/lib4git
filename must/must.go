package must

import (
	"context"
	"fmt"
	"runtime/debug"
)

type Error struct {
	Ctx   context.Context
	Stack []byte
	error
}

func mkErr(ctx context.Context, err error) Error {
	return Error{Ctx: ctx, Stack: debug.Stack(), error: err}
}

func Panic(ctx context.Context, err error) {
	panic(mkErr(ctx, err))
}

func Errorf(ctx context.Context, format string, args ...any) {
	Panic(ctx, fmt.Errorf(format, args...))
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
			err = r.(Error).error
		}
	}()
	f()
	return nil
}

func TryWithStack(f func()) (err error, stack []byte) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(Error).error
			stack = r.(Error).Stack
		}
	}()
	f()
	return nil, nil
}

func Try1[R1 any](f func() R1) (_ R1, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(Error).error
		}
	}()
	return f(), nil
}
