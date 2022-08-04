package utils

import (
	"context"
	"errors"

	"github.com/thejerf/suture/v4"
)

type (
	Agent[T any] struct {
		value T
		cmds  chan agentCmd[T]
	}

	agentResult[T any] struct {
		value T
		err   error
	}

	agentCmd[T any] interface {
		Execute(*Agent[T], context.Context)
	}

	updateCmd[T any] struct {
		update func(T) (T, error)
		result chan<- agentResult[T]
	}

	getCmd[T any] struct {
		result chan<- T
	}
)

var (
	_ suture.Service = (*Agent[any])(nil)
)

func NewAgent[T any](value T) *Agent[T] {
	return &Agent[T]{value, make(chan agentCmd[T], 10)}
}

func (a *Agent[T]) Serve(ctx context.Context) (err error) {
	for err == nil {
		select {
		case cmd, ok := <-a.cmds:
			if ok {
				cmd.Execute(a, ctx)
			} else {
				err = suture.ErrDoNotRestart
			}
		case <-ctx.Done():
			err = ctx.Err()
		}
	}
	return
}

func (a *Agent[T]) Update(update func(T) (T, error)) (T, error) {
	resultC := make(chan agentResult[T])
	a.cmds <- updateCmd[T]{update, resultC}
	result := <-resultC
	return result.value, result.err
}

func (a *Agent[T]) Get() (value T, err error) {
	resultC := make(chan T)
	a.cmds <- getCmd[T]{resultC}
	if result, ok := <-resultC; ok {
		value = result
	} else {
		err = errors.New("result channel closed unexpectedly")
	}
	return
}

func (c updateCmd[T]) Execute(a *Agent[T], ctx context.Context) {
	defer close(c.result)
	result := agentResult[T]{}
	result.value, result.err = c.update(a.value)
	select {
	case c.result <- result:
	case <-ctx.Done():
	}
}

func (c getCmd[T]) Execute(a *Agent[T], ctx context.Context) {
	defer close(c.result)
	select {
	case c.result <- a.value:
	case <-ctx.Done():
	}
}
