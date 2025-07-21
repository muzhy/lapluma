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
	error
}
