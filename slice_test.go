package lapluma_test

import (
	"lapluma"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceFilter(t *testing.T) {
	is := assert.New(t)

	s := []int{1, 2, 3}
	allIt := lapluma.FilterSlice(s, func(i int, v int) bool {
		return true
	})

	for _, value := range s {
		is.Equal(true, allIt.Next())
		is.Equal(value, allIt.Value())
	}
	is.Equal(false, allIt.Next())

	oddIt := lapluma.FilterSlice(s, func(int, v int) bool {
		return v%2 == 1
	})
	for oddIt.Next() {
		is.Equal(1, oddIt.Value()%2)
	}

	it2 := lapluma.FilterSlice([]string{"", "foo", "", "bar", ""}, func(_ int, v string) bool {
		return len(v) > 0
	})
	stringSlice := lapluma.Collect(it2)
	is.Equal(stringSlice, []string{"foo", "bar"})
}

func TestSliceReduce(t *testing.T) {
	is := assert.New(t)

	sum := lapluma.Reduce(
		lapluma.FilterSlice([]int{1, 2, 3, 4}, func(_ int, _ int) bool { return true }),
		func(init int, v int) int { return init + v },
		0,
	)
	is.Equal(10, sum)
}

func TestSliceReverse(t *testing.T) {
	is := assert.New(t)

	reserve := lapluma.Collect(
		lapluma.FilterSliceReverse([]int{1, 2, 3, 4}, func(_ int, _ int) bool { return true }),
	)
	is.Equal([]int{4, 3, 2, 1}, reserve)
}

func TestSliceFilterChan(t *testing.T) {
	is := assert.New(t)

	ch := lapluma.FilterSliceChan([]int{1, 2, 3, 4}, func(int, int) bool { return true })
	s := make([]int, 0, 4)
	for v := range ch {
		s = append(s, v)
	}
	is.Equal(s, []int{1, 2, 3, 4})

	ch2 := lapluma.FilterSliceReverseChan([]int{1, 2, 3, 4}, func(int, int) bool { return true })
	s2 := make([]int, 0, 4)
	for v := range ch2 {
		s2 = append(s2, v)
	}
	is.Equal(s2, []int{4, 3, 2, 1})
}
