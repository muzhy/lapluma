package pipe

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
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

// 测试并行 TryMap
func TestPipeTryMapParallel(t *testing.T) {
	ctx := context.Background()
	data := make([]string, 100)
	for i := range data {
		if i%10 == 0 {
			data[i] = "error"
		} else {
			data[i] = strconv.Itoa(i)
		}
	}

	p := FromSlice(data, ctx)

	// 使用原子计数器跟踪处理数量
	var processed uint64
	tryMappedPipe := TryMap(p, func(s string) (int, error) {
		atomic.AddUint64(&processed, 1)
		time.Sleep(1 * time.Millisecond) // 模拟处理延迟
		if s == "error" {
			return 0, fmt.Errorf("invalid input")
		}
		return strconv.Atoi(s)
	}, 10, 20) // 并行度10，缓冲区20

	result := Collect(tryMappedPipe)

	// 验证结果正确性（应跳过10个错误项）
	assert.Len(t, result, 90, "Should skip 10 error elements")

	// 验证并行处理
	assert.Greater(t, atomic.LoadUint64(&processed), uint64(50),
		"Parallel processing should handle more than 50 elements")
}

// 测试上下文取消时的并行处理
func TestPipeParallelWithCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	data := make([]int, 100)
	for i := range data {
		data[i] = i
	}

	p := FromSlice(data, ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	// 启动处理
	go func() {
		defer wg.Done()
		Collect(Filter(p, func(e int) bool {
			time.Sleep(10 * time.Millisecond)
			return true
		}, 5, 10))
	}()

	// 短暂延迟后取消上下文
	time.Sleep(5 * time.Millisecond)
	cancel()

	wg.Wait() // 等待goroutine退出

	// 验证管道已关闭
	_, open := <-p.inChan
	assert.False(t, open, "Channel should be closed after context cancellation")
}

// 获取当前 goroutine 的唯一 ID
func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	// 从堆栈信息中提取 goroutine ID
	id := uint64(0)
	for i := 10; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
		id = id*10 + uint64(b[i]-'0')
	}
	return id
}

// 测试并行 Filter 的 goroutine 数量
func TestPipeFilterParallelGoroutines(t *testing.T) {
	ctx := context.Background()
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	p := FromSlice(data, ctx)

	// 用于跟踪活跃的 goroutine
	var (
		activeGoroutines sync.Map
		maxConcurrent    int32
		currentCount     int32
		wg               sync.WaitGroup
	)

	// 设置并行度为 5
	const parallelism = 5

	filteredPipe := Filter(p, func(e int) bool {
		// 记录当前 goroutine
		goroutineID := getGoroutineID()
		activeGoroutines.Store(goroutineID, true)

		// 更新最大并发数
		current := atomic.AddInt32(&currentCount, 1)
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current > max {
				if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
					break
				}
			} else {
				break
			}
		}

		// 模拟工作负载
		time.Sleep(5 * time.Millisecond)

		atomic.AddInt32(&currentCount, -1)
		return e%2 == 0
	}, parallelism, 20) // 并行度5，缓冲区20

	wg.Add(1)
	go func() {
		defer wg.Done()
		result := Collect(filteredPipe)
		// 验证结果正确性
		assert.Len(t, result, 500, "Should filter out 50 even numbers")
	}()

	wg.Wait()

	// 统计实际使用的 goroutine 数量
	goroutineCount := 0
	activeGoroutines.Range(func(key, value interface{}) bool {
		goroutineCount++
		return true
	})

	// 验证结果
	assert.Equal(t, parallelism, goroutineCount,
		"Should use exactly %d goroutines", parallelism)

	assert.Equal(t, int32(parallelism), maxConcurrent,
		"Should reach %d concurrent executions", parallelism)
}

// 测试并行 Map 的处理顺序
func TestPipeMapParallelOrder(t *testing.T) {
	ctx := context.Background()
	data := make([]int, 100)
	for i := range data {
		data[i] = i
	}

	p := FromSlice(data, ctx)

	// 使用通道确保并行处理
	processingOrder := make(chan int, 100)

	// 设置并行度为 10
	mappedPipe := Map(p, func(e int) int {
		// 记录处理顺序
		processingOrder <- e
		time.Sleep(time.Duration(e%10) * time.Millisecond)
		return e
	}, 10, 20)

	result := Collect(mappedPipe)

	close(processingOrder)

	// 收集处理顺序
	var order []int
	for e := range processingOrder {
		order = append(order, e)
	}

	// 验证结果
	assert.Len(t, result, 100, "Should process all elements")

	// 验证处理顺序不是顺序的（证明并行处理）
	isSequential := true
	for i := range order {
		if i > 0 && order[i] < order[i-1] {
			isSequential = false
			break
		}
	}
	// 这里有小概率是顺序的,特别是当系统负载高导致没有发现并行时
	assert.False(t, isSequential,
		"Processing order should not be sequential in parallel execution")
}

// pipe_test.go
func TestBatch(t *testing.T) {
	t.Run("固定大小分批", func(t *testing.T) {
		ctx := context.Background()
		data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		p := FromSlice(data, ctx)

		batched := Batch(p, 3, time.Second)
		results := Collect(batched)

		expected := [][]int{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
			{10},
		}
		assert.Equal(t, expected, results)
	})

	t.Run("超时分批", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		input := make(chan int)
		p := Create(input, ctx)

		go func() {
			defer close(input)
			for i := 1; i <= 5; i++ {
				select {
				case input <- i:
					time.Sleep(300 * time.Millisecond)
				case <-ctx.Done():
					return
				}
			}
		}()

		batched := Batch(p, 10, 500*time.Millisecond)
		results := Collect(batched)

		// 预期：300ms * 2 = 600ms > 500ms
		// 应分成三批: [1,2], [3,4], [5]
		assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, results)
	})

	t.Run("上下文取消", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		input := make(chan int)
		p := Create(input, ctx)

		go func() {
			defer close(input)
			for i := 0; i < 100; i++ {
				select {
				case input <- i:
					time.Sleep(50 * time.Millisecond)
				case <-ctx.Done():
					return
				}
			}
		}()

		batched := Batch(p, 10, time.Second)
		go func() {
			results := Collect(batched)
			assert.True(t, len(results) < 10, "应提前终止")
		}()

		time.Sleep(time.Second)
		// 取消上下文
		cancel()
	})
}
