package lapluma

type SliceIterator[T any] struct {
	s         []T
	cursor    int
	step      int
	condition func(index int, value T) bool
}

func (sit *SliceIterator[T]) boundsCheck() bool {
	if sit.cursor >= len(sit.s) || sit.cursor < 0 {
		return false
	}
	return true
}

func (sit *SliceIterator[T]) Next() bool {
	if !sit.boundsCheck() {
		return false
	}
	for !sit.condition(sit.cursor, sit.s[sit.cursor]) {
		sit.cursor += sit.step
		if !sit.boundsCheck() {
			return false
		}
	}
	return true
}

func (sit *SliceIterator[T]) Value() (v T) {
	v = sit.s[sit.cursor]
	sit.cursor += sit.step
	return
}

func trueFilter[T any](int, T) bool {
	return true
}

func FilterSlice[T any](s []T, handler func(index int, value T) bool) Iterator[T] {
	return &SliceIterator[T]{
		cursor:    0,
		step:      1,
		s:         s,
		condition: handler,
	}
}

func ForeachSlice[T any](s []T) Iterator[T] {
	return FilterSlice(s, trueFilter)
}

func FilterSliceReverse[T any](s []T, handler func(index int, value T) bool) Iterator[T] {
	return &SliceIterator[T]{
		cursor:    len(s) - 1,
		step:      -1,
		s:         s,
		condition: handler,
	}
}

func ForeachSliceReverse[T any](s []T) Iterator[T] {
	return FilterSliceReverse(s, trueFilter)
}

func FilterSliceChan[T any](s []T, handler func(index int, value T) bool, size ...int) <-chan T {
	buffSize := 0
	if len(size) > 0 {
		buffSize = size[0]
	}
	ch := make(chan T, buffSize)
	go func() {
		defer close(ch)
		for i, v := range s {
			if handler(i, v) {
				ch <- v
			}
		}
	}()
	return ch
}

func ForeachSliceChan[T any](s []T, handler func(index int, value T) bool, size ...int) <-chan T {
	return FilterSliceChan(s, trueFilter, size...)
}

func FilterSliceReverseChan[T any](s []T, handler func(index int, value T) bool, size ...int) <-chan T {
	buffSize := 0
	if len(size) > 0 {
		buffSize = size[0]
	}
	ch := make(chan T, buffSize)
	go func() {
		defer close(ch)
		for i := len(s) - 1; i >= 0; i-- {
			if handler(i, s[i]) {
				ch <- s[i]
			}
		}
	}()
	return ch
}

func ForeachSliceReverseChan[T any](s []T, handler func(index int, value T) bool, size ...int) <-chan T {
	return FilterSliceReverseChan(s, trueFilter, size...)
}
