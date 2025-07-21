package iterator_test

import (
	"lapluma/iterator"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceIterator(t *testing.T) {
	is := assert.New(t)

	data := []int{1, 2, 3}
	it := iterator.FromSlice(data)

	for _, v := range data {
		e, ok := it.Next()
		is.Equal(v, e)
		is.Equal(true, ok)
	}

	v, ok := it.Next()
	is.Equal(0, v)
	is.Equal(false, ok)
}

func TestMapIterator(t *testing.T) {
	is := assert.New(t)

	data := make(map[string]string)
	data["a"] = "apple"
	data["b"] = "banana"

	it := iterator.FromMap(data)
	for k, v := range iterator.Iter2(it) {
		is.Equal(data[k], v)
	}

	keysIt := iterator.MapKeysIt(data)
	k, ok := keysIt.Next()
	is.Equal(true, ok)
	is.Equal(k, "a")
}

func TestFilterIterator(t *testing.T) {
	is := assert.New(t)

	data := []int{1, 2, 3}
	it := iterator.Filter(
		iterator.FromSlice(data),
		func(n int) bool { return n%2 == 1 },
	)
	for num := range iterator.Iter(it) {
		is.Equal(1, num%2)
	}

	sum := iterator.Reduce(
		iterator.FromSlice(data),
		func(init int, e int) int { return init + e },
		0,
	)
	is.Equal(6, sum)

	doubleIt := iterator.Map(
		iterator.FromSlice(data),
		func(n int) int {
			return n * 2
		},
	)
	douleSum := iterator.Reduce(
		doubleIt,
		func(init int, e int) int { return init + e },
		0,
	)
	is.Equal(sum*2, douleSum)
}

func TestGroup(t *testing.T) {
	is := assert.New(t)

	data := []int{1, 2, 3, 4}

	m := iterator.Group(
		iterator.FromSlice(data),
		func(n int) (string, int) {
			if n%2 == 1 {
				return "odd", n
			} else {
				return "even", n
			}
		},
	)

	is.Equal([]int{1, 3}, m["odd"])
	is.Equal([]int{2, 4}, m["even"])
	is.Equal(2, len(m))
}
