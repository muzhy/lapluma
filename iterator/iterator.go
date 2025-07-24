package iterator

import (
	"iter"

	"github.com/muzhy/lapluma"
)

type Iterator[E any] interface {
	// !!! 在任何转换迭代器的 Next() 方法内部，严禁使用 for-range 模式从上游迭代器获取数据。必须使用传统的、最高效的方式
	Next() (E, bool)
}

// 1.23之后，标准库中增加迭代器`iter`
// 通过`Iter()`创建符合标准库格式的迭代器`iter.Seq[E]`
// 可以直接使用rang进行遍历
func Iter[E any](it Iterator[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		for data, ok := it.Next(); ok; data, ok = it.Next() {
			if !yield(data) {
				return
			}
		}
	}
}

func Iter2[K, V any](it Iterator[lapluma.Pair[K, V]]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for data, ok := it.Next(); ok; data, ok = it.Next() {
			k, v := data.Extra()
			if !yield(k, v) {
				return
			}
		}
	}
}

type BaseIterator[E any] struct {
	nextFunc func() (E, bool)
}

func (it *BaseIterator[E]) Next() (E, bool) {
	return it.nextFunc()
}

// support function to help create iterator from slice and map
// Iterator 用于串行处理场景，不支持并发，如果需要在并发场景下使用
// 调用pipe.`FromIterator`将其转换为Pipe
func FromSlice[E any](data []E) Iterator[E] {
	cursor := 0
	// var mut sync.Mutex
	next := func() (E, bool) {
		// mut.Lock()
		// defer mut.Unlock()
		if cursor < len(data) {
			index := cursor
			cursor++
			return data[index], true
		} else {
			var zero E
			return zero, false
		}
	}
	return &BaseIterator[E]{
		nextFunc: next,
	}
}

func FromMap[K comparable, V any](data map[K]V) Iterator[lapluma.Pair[K, V]] {
	// keyIt := MapKeysIt(data)

	keys := make([]K, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	cursor := 0

	next := func() (lapluma.Pair[K, V], bool) {
		if cursor < len(keys) {
			key := keys[cursor]
			cursor++
			return lapluma.Pair[K, V]{
				First:  key,
				Second: data[key],
			}, true
		}
		var zero lapluma.Pair[K, V]
		return zero, false
	}

	return &BaseIterator[lapluma.Pair[K, V]]{
		nextFunc: next,
	}
}

type FilterIterator[E any] struct {
	input  Iterator[E]
	filter func(E) bool
}

func (it *FilterIterator[E]) Next() (E, bool) {
	for e, ok := it.input.Next(); ok; e, ok = it.input.Next() {
		// for e := range Iter(it.input) {
		if it.filter(e) {
			return e, true
		}
	}
	var zero E
	return zero, false
}

func Filter[E any](it Iterator[E], filter func(E) bool) Iterator[E] {
	return &FilterIterator[E]{
		input:  it,
		filter: filter,
	}
}

// `Collect`和`Group`理论上可以使用`Map`+`Reduce`实现
// 提供这两个函数主要是为了更方便使用
func Collect[E any](it Iterator[E]) []E {
	s := make([]E, 0)
	// for e, ok := it.Next(); ok; e, ok = it.Next() {
	for e := range Iter(it) {
		s = append(s, e)
	}
	return s
}

// `Group` and `Map` not support error handler
// if source data have some invalid value, should use `Filter` to filter data before
// call `Group“ or `Map` or `Reduce`
// if some external resources faile, `extract` or `handler` should panic or return
// some default value
func Group[K comparable, E, R any](it Iterator[E], extract func(E) (K, R)) map[K][]R {
	m := make(map[K][]R)
	for e := range Iter(it) {
		k, r := extract(e)
		m[k] = append(m[k], r)
	}
	return m
}

type MapIterator[E, R any] struct {
	handler func(E) R
	input   Iterator[E]
	closed  bool
}

func (it *MapIterator[E, R]) stop() (R, bool) {
	it.closed = true
	var zero R
	return zero, false
}

func (it *MapIterator[E, R]) Next() (R, bool) {
	if it.closed {
		return it.stop()
	}
	if e, ok := it.input.Next(); ok {
		// for e := range Iter(it.input) {
		return it.handler(e), true
	}
	return it.stop()
}

// 这里是否可以不要errHandler，由调用者负责保证`hanlder`的正确性
// 但是如果确实出现了特殊情况handler无法处理，由必须提供手段给调用者
func Map[E, R any](it Iterator[E], handler func(E) R) Iterator[R] {
	return &MapIterator[E, R]{
		input:   it,
		handler: handler,
		closed:  false,
	}
}

// TryMapIterator 用于处理那些可能会发生错误的操作
type TryMapIterator[E, R any] struct {
	handler func(E) (R, error)
	input   Iterator[E]
}

func (it *TryMapIterator[E, R]) Next() (R, bool) {
	for e, ok := it.input.Next(); ok; e, ok = it.input.Next() {
		r, err := it.handler(e)
		if err != nil {
			// On error, skip this element and try the next one.
			continue
		}
		return r, true
	}
	var zero R
	return zero, false
}

func TryMap[E, R any](it Iterator[E], handler func(E) (R, error)) Iterator[R] {
	return &TryMapIterator[E, R]{
		input:   it,
		handler: handler,
	}
}

func Reduce[E, R any](it Iterator[E], handler func(R, E) R, initial R) R {
	// for e, ok := it.Next(); ok; e, ok = it.Next() {
	for e := range Iter(it) {
		initial = handler(initial, e)
	}
	return initial
}
