package utils

import (
	"context"
	"sync"

	log "github.com/inconshreveable/log15"
	"github.com/thejerf/suture/v4"
)

type (
	Dispatcher[T any] struct {
		*Agent[subscribers[T]]
		NewSubscriberHook func(chan<- T)
	}

	subscribers[T any] [](chan T)
)

var (
	_ suture.Service = (*Dispatcher[any])(nil)

	Log = log.New()
)

func NewDispatcher[T any]() *Dispatcher[T] {
	return &Dispatcher[T]{Agent: NewAgent[subscribers[T]](nil)}
}

func (d *Dispatcher[T]) OnNewSubscriber(hook func(chan<- T)) {
	d.NewSubscriberHook = hook
}

func (d *Dispatcher[T]) Subscribe() (c <-chan T, cancel func()) {
	bidiChan := make(chan T)
	c = bidiChan
	_, _ = d.Agent.Update(func(subs subscribers[T]) (subscribers[T], error) {
		subs = append(subs, bidiChan)
		Log.Debug("added subscriber", "c", bidiChan)
		return subs, nil
	})
	if d.NewSubscriberHook != nil {
		go d.NewSubscriberHook(bidiChan)
	}
	cancel = func() {
		_, _ = d.Agent.Update(func(subs subscribers[T]) (subscribers[T], error) {
			j := 0
			for i, sub := range subs {
				if sub != bidiChan {
					subs[j] = subs[i]
					j++
				}
			}
			Log.Debug("removed subscriber", "c", c)
			close(bidiChan)
			return subs[:j], nil
		})
	}
	return
}

func (d *Dispatcher[T]) Dispatch(value T, ctx context.Context) (err error) {
	subs, err := d.Agent.Get()
	if err != nil {
		return err
	}
	Log.Debug("dispatching event", "event", value, "#sub", len(subs))
	if len(subs) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(subs))
	for _, target := range subs {
		go func(target chan<- T) {
			defer wg.Done()
			select {
			case target <- value:
			case <-ctx.Done():
				err = ctx.Err()
				return
			}
		}(target)
	}
	wg.Wait()
	return
}
