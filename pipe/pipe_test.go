package pipe

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"lapluma/iterator"
	// "lapluma/pipe"

	"github.com/stretchr/testify/assert"
)

// 测试从切片创建管道
func TestFromSlice(t *testing.T) {
	ctx := context.Background()
	data := []int{1, 2, 3}
	p := FromSlice(data, ctx)

	// 从管道中收集结果进行验证
	result := Collect(p)

	assert.Equal(t, data, result, "FromSlice should create a pipe that sends all elements of the slice")
}

// 测试从迭代器创建管道
func TestFromIterator(t *testing.T) {
	ctx := context.Background()
	data := []int{1, 2, 3}
	it := iterator.FromSlice(data)
	p := FromIterator(it, ctx)

	result := Collect(p)

	assert.Equal(t, data, result, "FromIterator should create a pipe from an iterator")
}

// 测试管道的 Filter 操作
func TestPipeFilter(t *testing.T) {
	ctx := context.Background()
	data := []int{1, 2, 3, 4, 5, 6}
	p := FromSlice(data, ctx)

	// 过滤偶数
	filteredPipe := Filter(p, func(e int) bool {
		return e%2 == 0
	})

	result := Collect(filteredPipe)

	expected := []int{2, 4, 6}
	assert.Equal(t, expected, result, "Pipe Filter should correctly filter elements")
}

// 测试管道的 Map 操作
func TestPipeMap(t *testing.T) {
	ctx := context.Background()
	data := []int{1, 2, 3}
	p := FromSlice(data, ctx)

	// 将整数转换为字符串
	mappedPipe := Map(p, func(e int) string {
		return "item:" + strconv.Itoa(e)
	})

	result := Collect(mappedPipe)

	expected := []string{"item:1", "item:2", "item:3"}
	assert.Equal(t, expected, result, "Pipe Map should correctly transform elements")
}

// 测试管道的 Reduce 操作
func TestPipeReduce(t *testing.T) {
	ctx := context.Background()
	data := []int{1, 2, 3, 4}
	p := FromSlice(data, ctx)

	// 计算总和
	sum := Reduce(p, func(acc int, e int) int {
		return acc + e
	}, 10)

	assert.Equal(t, 20, sum, "Pipe Reduce should correctly accumulate values")
}

// 测试管道的 Group 操作
func TestPipeGroup(t *testing.T) {
	ctx := context.Background()
	type Product struct {
		Name     string
		Category string
	}
	data := []Product{
		{Name: "Apple", Category: "Fruit"},
		{Name: "Carrot", Category: "Vegetable"},
		{Name: "Banana", Category: "Fruit"},
	}
	p := FromSlice(data, ctx)

	// 按类别分组
	grouped := Group(p, func(p Product) (string, string) {
		return p.Category, p.Name
	})

	expected := map[string][]string{
		"Fruit":     {"Apple", "Banana"},
		"Vegetable": {"Carrot"},
	}

	// 对比 map 时，需要确保切片的顺序一致
	sort.Strings(grouped["Fruit"])
	sort.Strings(grouped["Vegetable"])

	assert.True(t, reflect.DeepEqual(expected, grouped), "Pipe Group should group items correctly")
}

// 测试上下文取消
func TestPipeWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	data := []int{1, 2, 3, 4, 5}
	p := FromSlice(data, ctx)

	// 读取一个元素后取消上下文
	<-p.inChan
	cancel()

	// 等待一小段时间以确保 goroutine 有时间响应取消信号
	time.Sleep(10 * time.Millisecond)

	// 检查管道是否已关闭
	_, ok := <-p.inChan
	assert.False(t, ok, "Pipe should be closed after context cancellation")
}

// 测试链式操作
func TestPipeChaining(t *testing.T) {
	ctx := context.Background()
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// 链式操作：过滤 -> 映射 -> 过滤
	p := FromSlice(data, ctx)
	filteredPipe := Filter(p, func(e int) bool {
		return e > 3 // {4, 5, 6, 7, 8, 9, 10}
	})
	mappedPipe := Map(filteredPipe, func(e int) string {
		return "val" + strconv.Itoa(e) // {"val4", ..., "val10"}
	})
	finalPipe := Filter(mappedPipe, func(s string) bool {
		return len(s) > 4 // {"val10"}
	})

	result := Reduce(finalPipe, func(acc []string, e string) []string {
		return append(acc, e)
	}, []string{})

	expected := []string{"val10"}
	assert.Equal(t, expected, result, "Pipe chaining should work correctly")
}
