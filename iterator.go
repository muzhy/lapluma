package lapluma

type Iterator[T any] interface {
	Next() bool
	Value() T
}

func Collect[E any](it Iterator[E]) []E {
	s := make([]E, 0)
	for it.Next() {
		s = append(s, it.Value())
	}
	return s
}

func Group[K comparable, E any](it Iterator[E], getKey func(e E) K) map[K][]E {
	m := make(map[K][]E)
	for it.Next() {
		e := it.Value()
		key := getKey(e)
		m[key] = append(m[key], e)
	}
	return m
}

type mIterator[E, R any] struct {
	it      Iterator[E]
	handler func(E) R
}

func (mit *mIterator[E, R]) Next() bool {
	return mit.it.Next()
}

func (mit *mIterator[E, R]) Value() R {
	return mit.handler(mit.it.Value())
}

func Map[E, R any](it Iterator[E], handler func(E) R) Iterator[R] {
	return &mIterator[E, R]{
		it:      it,
		handler: handler,
	}
}

func Reduce[E, R any](it Iterator[E], handler func(R, E) R, initial R) R {
	for it.Next() {
		initial = handler(initial, it.Value())
	}
	return initial
}
