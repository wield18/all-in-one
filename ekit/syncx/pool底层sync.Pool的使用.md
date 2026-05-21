`sync.Pool` 是 Go 语言标准库中用于**临时对象池**的工具，主要目的是**减少内存分配和 GC 压力**。

## 核心概念

```go
type Pool struct {
    New func() interface{}  // 当池中没有对象时，调用此函数创建
    // 内部实现：每个 P 有本地缓存，无锁或轻量锁
}
```

## 基本用法

### 1. 最简单的使用

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    // 创建对象池
    pool := sync.Pool{
        New: func() interface{} {
            fmt.Println("创建新对象")
            return &[]int{}
        },
    }
    
    // 从池中获取对象
    obj := pool.Get().(*[]int)
    fmt.Printf("获取对象: %p\n", obj)
    
    // 使用对象
    *obj = append(*obj, 1, 2, 3)
    
    // 放回池中
    pool.Put(obj)
    
    // 再次获取（复用之前的对象）
    obj2 := pool.Get().(*[]int)
    fmt.Printf("再次获取: %p\n", obj2)  // 相同地址
    fmt.Println(*obj2)  // 输出: [1 2 3]
}
```

### 2. 泛型封装（你的代码就是这种）

```go
type Pool[T any] struct {
    p sync.Pool
}

func NewPool[T any](new func() T) *Pool[T] {
    return &Pool[T]{
        p: sync.Pool{
            New: func() interface{} {
                return new()
            },
        },
    }
}

func (p *Pool[T]) Get() T {
    return p.p.Get().(T)
}

func (p *Pool[T]) Put(x T) {
    p.p.Put(x)
}

// 使用
type User struct {
    ID   int
    Name string
}

userPool := NewPool(func() *User {
    return &User{}
})

user := userPool.Get()  // 类型是 *User，无需断言
defer userPool.Put(user)
```

## 典型应用场景

### 场景1：缓冲区复用（最经典）

```go
// 高并发日志系统
var bufferPool = sync.Pool{
    New: func() interface{} {
        return &bytes.Buffer{}
    },
}

func processLog(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    
    buf.Reset()  // 重要：清空之前的数据
    buf.Write(data)
    
    // 使用 buf...
}

// 性能对比：
// 无Pool: 每次创建新Buffer，大量内存分配
// 有Pool: 复用Buffer，内存分配减少90%+
```

### 场景2：JSON编解码

```go
var encoderPool = sync.Pool{
    New: func() interface{} {
        return json.NewEncoder(&bytes.Buffer{})
    },
}

func encodeData(data interface{}) ([]byte, error) {
    buf := &bytes.Buffer{}
    encoder := json.NewEncoder(buf)
    
    if err := encoder.Encode(data); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

// 更优写法（复用encoder）
var bufferPool = sync.Pool{
    New: func() interface{} {
        return &bytes.Buffer{}
    },
}

func encodeWithPool(data interface{}) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()
    
    if err := json.NewEncoder(buf).Encode(data); err != nil {
        return nil, err
    }
    
    // 复制结果，因为 buf 会被复用
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes())
    return result, nil
}
```

### 场景3：数据库连接/事务

```go
type Transaction struct {
    tx *sql.Tx
    // 其他字段
}

var txPool = sync.Pool{
    New: func() interface{} {
        return &Transaction{}
    },
}

func getTransaction() *Transaction {
    return txPool.Get().(*Transaction)
}

func putTransaction(tx *Transaction) {
    tx.tx = nil  // 清理引用
    txPool.Put(tx)
}
```

## 重要特性（容易踩坑）

### 1. 对象可能被自动清除

```go
pool := sync.Pool{New: func() interface{} { return &[]int{} }}

obj := pool.Get().(*[]int)
pool.Put(obj)

// GC 发生后，对象可能被清除
runtime.GC()

obj2 := pool.Get().(*[]int)
fmt.Println(obj == obj2)  // 可能 false，对象被重建了
```

### 2. Put 前必须重置状态

```go
// ❌ 错误：没有重置
buf := bufferPool.Get().(*bytes.Buffer)
buf.WriteString("hello")
bufferPool.Put(buf)  // 下次获取时还包含 "hello"

// ✅ 正确：重置状态
buf := bufferPool.Get().(*bytes.Buffer)
buf.WriteString("hello")
buf.Reset()  // 清空
bufferPool.Put(buf)
```

### 3. Get 后必须类型断言

```go
// ❌ 错误：忘记类型断言
obj := pool.Get()  // interface{}
obj.Field = 1      // 编译错误

// ✅ 正确：类型断言
obj := pool.Get().(*MyStruct)
obj.Field = 1
```

### 4. 不要存储需要关闭的资源

```go
// ❌ 错误：存储需要关闭的连接
connPool.Put(dbConn)  // 如果 conn 被清理，不会自动关闭

// ✅ 正确：在 Put 前关闭，或在 New 中创建新连接
var connPool = sync.Pool{
    New: func() interface{} {
        conn, _ := sql.Open("mysql", dsn)
        return conn
    },
}
```

## 性能测试示例

```go
// 对比有 Pool 和无 Pool 的性能
func BenchmarkWithoutPool(b *testing.B) {
    for i := 0; i < b.N; i++ {
        buf := &bytes.Buffer{}
        buf.WriteString("hello")
        _ = buf.String()
    }
}

func BenchmarkWithPool(b *testing.B) {
    pool := sync.Pool{
        New: func() interface{} {
            return &bytes.Buffer{}
        },
    }
    
    for i := 0; i < b.N; i++ {
        buf := pool.Get().(*bytes.Buffer)
        buf.WriteString("hello")
        _ = buf.String()
        buf.Reset()
        pool.Put(buf)
    }
}

// 典型结果：
// BenchmarkWithoutPool-8    5000000    280 ns/op    48 B/op    2 allocs/op
// BenchmarkWithPool-8      20000000     85 ns/op     0 B/op    0 allocs/op
// 性能提升 3倍+，零内存分配！
```

## 你的 `Pool[T any]` 的优势

```go
// 标准库（需要类型断言）
var stdPool = sync.Pool{New: func() interface{} { return &User{} }}
user := stdPool.Get().(*User)  // 麻烦

// 你的封装（类型安全）
type Pool[T any] struct { p sync.Pool }
userPool := NewPool(func() *User { return &User{} })
user := userPool.Get()  // 自动是 *User 类型，无需断言
```

## 最佳实践总结

1. **适用场景**：频繁创建和销毁的临时对象（缓冲区、编解码器、临时结构体）
2. **不适用场景**：数据库连接池、长期持有的对象
3. **必须重置**：Put 前清空对象状态
4. **不要假设**：对象会一直保留（随时可能被 GC）
5. **合理设置 New**：必须设置，否则 Get 可能返回 nil
6. **注意并发**：`sync.Pool` 本身是并发安全的

这是一个**性能优化的利器**，但不是万能的，要理解其"临时性"的特点！