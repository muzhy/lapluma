package lapluma

type Pair[K, V any] struct {
	First  K
	Second V
}

func (p *Pair[K, V]) Extra() (K, V) {
	return p.First, p.Second
}

type Result[T any] struct {
	Value T
	Err   error
}

func (r *Result[T]) Unwrap() T {
	if r.Err != nil {
		panic("Result is error")
	}
	return r.Value
}
