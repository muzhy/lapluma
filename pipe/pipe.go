package pipe

import (
	"context"
	"lapluma/iterator"
)

type Pipe[E any] struct {
	inChan <-chan E
	ctx    context.Context
}

func Create[E any](inChan <-chan E, ctx context.Context) *Pipe[E] {
	return &Pipe[E]{
		inChan: inChan,
		ctx:    ctx,
	}
}

func FromSlice[E any](data []E, ctx context.Context) *Pipe[E] {
	inChan := make(chan E)

	go func() {
		defer close(inChan)
		for _, d := range data {
			select {
			case inChan <- d:
			case <-ctx.Done():
				return
			}
		}
	}()

	return Create(inChan, ctx)
}

func FromIterator[E any](it iterator.Iterator[E], ctx context.Context) *Pipe[E] {
	inChan := make(chan E)

	go func() {
		defer close(inChan)
		for data, ok := it.Next(); ok; data, ok = it.Next() {
			select {
			case inChan <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	return Create(inChan, ctx)
}

func Filter[E any](input *Pipe[E], filter func(E) bool) *Pipe[E] {
	output := make(chan E)

	go func() {
		defer close(output)
		for {
			select {
			case <-input.ctx.Done():
				return
			case data, ok := <-input.inChan:
				if !ok {
					return
				}
				if filter(data) {
					select {
					case <-input.ctx.Done():
						return
					case output <- data:
					}
				} else {
					continue
				}
			}
		}
	}()

	return &Pipe[E]{
		inChan: output,
		ctx:    input.ctx,
	}
}

func Map[T, R any](inPipe *Pipe[T], trans func(T) R) *Pipe[R] {
	outChan := make(chan R)

	go func() {
		defer close(outChan)
		for {
			select {
			case <-inPipe.ctx.Done():
				return
			case data, ok := <-inPipe.inChan:
				if !ok {
					return // channel closed
				}
				r := trans(data)

				select {
				case outChan <- r:
				case <-inPipe.ctx.Done():
					return
				}
			}
		}
	}()

	return &Pipe[R]{
		inChan: outChan,
		ctx:    inPipe.ctx,
	}
}

func Reduce[E, R any](input *Pipe[E], handler func(R, E) R, initial R) R {
	for data := range input.inChan {
		initial = handler(initial, data)
	}
	return initial
}

func Group[K comparable, E, R any](inPipe *Pipe[E], extract func(E) (K, R)) map[K][]R {
	m := make(map[K][]R)
	for data := range inPipe.inChan {
		k, r := extract(data)
		m[k] = append(m[k], r)
	}
	return m
}

func Collect[E any](inPipe *Pipe[E]) []E {
	var result []E
	for item := range inPipe.inChan {
		result = append(result, item)
	}
	return result
}
