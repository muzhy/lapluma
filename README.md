# lapluma
支持两种模式`Iterator`和`Pipe`，`Pipe`主要是通过`channel`实现，为这两种模式提供`Filter`，`Map`，`Reduce`函数

`Map`，`Reduce` 不提供错误处理，简化设计和调用
如果数据存在非法数据，，应先通过`Filter`将此类数据剔除，
如果是在处理单个数据的过程中可能由于外部资源或其他问题导致处理失败，可以将`handler`的返回值设置为`Result[E, error]`类型，在`Map`之后在添加一个`Filter`将失败数据过滤掉，或者直接在`Reduce`的`handler`中忽略该数据。

提供`Iter()`和`Iter2()`函数，从`Iterator`中返回符合1.23新增的`iter.Seq[E]`和`iter.Seq2[K, V]`的迭代器，支持通过`for-rang`遍历迭代器