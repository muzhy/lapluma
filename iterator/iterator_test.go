package iterator

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromSlice(t *testing.T) {
	data := []int{1, 2, 3}
	it := FromSlice(data)

	result := Collect(it)
	assert.Equal(t, data, result, "FromSlice should create an iterator that yields all elements of the slice")

	// 测试空切片
	itEmpty := FromSlice([]int{})
	resultEmpty := Collect(itEmpty)
	assert.Empty(t, resultEmpty, "FromSlice with an empty slice should result in an empty iterator")
}

// 测试从 Map 创建迭代器
func TestFromMap(t *testing.T) {
	data := map[string]int{"a": 1, "b": 2}
	it := FromMap(data)

	// 从迭代器收集结果
	pairs := Collect(it)

	// 验证结果
	assert.Equal(t, 2, len(pairs), "FromMap should create an iterator with the same number of elements as the map")

	// 将结果转回 map 以便验证，因为 map 的迭代顺序是不确定的
	resultMap := make(map[string]int)
	for _, p := range pairs {
		resultMap[p.First] = p.Second
	}
	assert.Equal(t, data, resultMap, "The collected map should be equal to the original map")
}

// 测试 Filter 函数
func TestFilter(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6}
	it := FromSlice(data)

	// 过滤出偶数
	filteredIt := Filter(it, func(e int) bool {
		return e%2 == 0
	})

	result := Collect(filteredIt)
	expected := []int{2, 4, 6}
	assert.Equal(t, expected, result, "Filter should correctly filter elements based on the predicate")
}

// 测试 Map 函数
func TestMap(t *testing.T) {
	data := []int{1, 2, 3}
	it := FromSlice(data)

	// 将整数转换为字符串
	mappedIt := Map(it, func(e int) string {
		return "v" + strconv.Itoa(e)
	})

	result := Collect(mappedIt)
	expected := []string{"v1", "v2", "v3"}
	assert.Equal(t, expected, result, "Map should correctly transform each element")
}

// 测试 Reduce 函数
func TestReduce(t *testing.T) {
	data := []int{1, 2, 3, 4}
	it := FromSlice(data)

	// 计算总和
	sum := Reduce(it, func(acc int, e int) int {
		return acc + e
	}, 0)

	assert.Equal(t, 10, sum, "Reduce should correctly accumulate the values")

	// 测试初始值
	sumWithInitial := Reduce(FromSlice(data), func(acc int, e int) int {
		return acc + e
	}, 10)
	assert.Equal(t, 20, sumWithInitial, "Reduce should work correctly with a non-zero initial value")
}

// 测试 Group 函数
func TestGroup(t *testing.T) {
	type Person struct {
		Name string
		City string
	}
	data := []Person{
		{Name: "Alice", City: "New York"},
		{Name: "Bob", City: "Tokyo"},
		{Name: "Charlie", City: "New York"},
	}
	it := FromSlice(data)

	// 按城市分组
	grouped := Group(it, func(p Person) (string, string) {
		return p.City, p.Name
	})

	expected := map[string][]string{
		"New York": {"Alice", "Charlie"},
		"Tokyo":    {"Bob"},
	}

	// 使用 reflect.DeepEqual 比较 map，因为 assert.Equal 对 map 中的切片顺序敏感
	if !reflect.DeepEqual(expected, grouped) {
		t.Errorf("Group() = %v, want %v", grouped, expected)
	}
}

// 测试 Collect 函数
func TestCollect(t *testing.T) {
	data := []int{10, 20, 30}
	it := FromSlice(data)
	result := Collect(it)
	assert.Equal(t, data, result, "Collect should gather all iterator elements into a slice")
}
