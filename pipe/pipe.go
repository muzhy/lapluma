package pipe

import (
	"context"
	"lapluma/iterator"
	"sync"
	"time"
)

type Pipe[E any] struct {
	inChan <-chan E
	ctx    context.Context
}

func (p *Pipe[E]) Next() (E, bool) {
	var zero E
	if p.inChan == nil {
		return zero, false
	}
	select {
	case <-p.ctx.Done():
		return zero, false
	case data, ok := <-p.inChan:
		return data, ok
	}
}

func Create[E any](inChan <-chan E, ctx context.Context) *Pipe[E] {
	return &Pipe[E]{
		inChan: inChan,
		ctx:    ctx,
	}
}

func FromSlice[E any](data []E, ctx context.Context) *Pipe[E] {
	output := make(chan E)

	go func() {
		defer close(output)
		for _, d := range data {
			select {
			case output <- d:
			case <-ctx.Done():
				return
			}
		}
	}()

	return Create(output, ctx)
}

func FromIterator[E any](it iterator.Iterator[E], ctx context.Context) *Pipe[E] {
	output := make(chan E)

	go func() {
		defer close(output)
		for data := range iterator.Iter(it) {
			select {
			case output <- data:
			case <-ctx.Done():
				return
			}
		}
	}()

	return Create(output, ctx)
}

func parallel[E any](executer func(wg *sync.WaitGroup, output chan E), paralleism int, bufSize int) <-chan E {
	var wg sync.WaitGroup
	output := make(chan E, bufSize)

	go func() {
		defer close(output)
		defer wg.Wait()
		for i := 0; i < paralleism; i++ {
			wg.Add(1)
			go executer(&wg, output)
		}
	}()

	return output
}

func extractParallelParam(size ...int) (int, int) {
	paralleism := 1
	if len(size) > 0 && size[0] > 0 {
		paralleism = size[0]
	}

	outbuf := 0
	if len(size) > 1 && size[1] > 0 {
		outbuf = size[1]
	}

	return paralleism, outbuf
}

// `Filter`, `Map` 默认不进行并行，只启动一个goruntine
func Filter[E any](input *Pipe[E], filter func(E) bool, size ...int) *Pipe[E] {
	paralleism, buf := extractParallelParam(size...)

	output := parallel(
		func(wg *sync.WaitGroup, output chan E) {
			defer wg.Done()
			for data := range iterator.Iter(input) {
				if filter(data) {
					select {
					case <-input.ctx.Done():
						return
					case output <- data:
					}
				}
			}
		},
		paralleism, buf,
	)

	return Create(output, input.ctx)
}

func Map[E, R any](inPipe *Pipe[E], trans func(E) R, size ...int) *Pipe[R] {
	paralleism, buf := extractParallelParam(size...)

	output := parallel(func(wg *sync.WaitGroup, output chan R) {
		defer wg.Done()
		for data := range iterator.Iter(inPipe) {
			r := trans(data)
			select {
			case output <- r:
			case <-inPipe.ctx.Done():
				return
			}
		}
	}, paralleism, buf)

	return Create(output, inPipe.ctx)
}

func TryMap[T, R any](inPipe *Pipe[T], trans func(T) (R, error), size ...int) *Pipe[R] {
	paralleism, buf := extractParallelParam(size...)
	output := parallel(func(wg *sync.WaitGroup, output chan R) {
		defer wg.Done()
		for data := range iterator.Iter(inPipe) {
			r, err := trans(data)
			if err != nil {
				continue
			}
			select {
			case output <- r:
			case <-inPipe.ctx.Done():
				return
			}
		}
	}, paralleism, buf)

	return Create(output, inPipe.ctx)
}

func Reduce[E, R any](input *Pipe[E], handler func(R, E) R, initial R) R {
	for data := range iterator.Iter(input) {
		initial = handler(initial, data)
	}
	return initial
}

func Group[K comparable, E, R any](input *Pipe[E], extract func(E) (K, R)) map[K][]R {
	m := make(map[K][]R)
	for data := range iterator.Iter(input) {
		k, r := extract(data)
		m[k] = append(m[k], r)
	}
	return m
}

func Collect[E any](input *Pipe[E]) []E {
	result := make([]E, 0)
	for data := range iterator.Iter(input) {
		result = append(result, data)
	}
	return result
}

func Batch[E any](input *Pipe[E], batchSize int, timeout time.Duration) *Pipe[[]E] {
	output := make(chan []E)

	go func() {
		defer close(output)
		buffer := make([]E, 0, batchSize)
		timer := time.NewTimer(timeout)

		for {
			select {
			case <-input.ctx.Done():
				return
			case data, ok := <-input.inChan:
				if !ok {
					if len(buffer) > 0 {
						// output <- buffer
						select {
						case output <- buffer:
						case <-input.ctx.Done():
							return
						}
					}
					return
				}

				buffer = append(buffer, data)
				if len(buffer) == batchSize {
					// output <- buffer
					select {
					case output <- buffer:
					case <-input.ctx.Done():
						return
					}
					buffer = make([]E, 0, batchSize)
					timer.Reset(timeout)
				}

			case <-timer.C:
				if len(buffer) > 0 {
					// output <- buffer
					select {
					case output <- buffer:
					case <-input.ctx.Done():
						return
					}
					buffer = make([]E, 0, batchSize)
				}
				timer.Reset(timeout)
			}
		}
	}()

	return Create(output, input.ctx)
}
