这是一个**分段锁（Segmented Lock）**实现，用于减少锁竞争。它将key分散到多个锁上，不同的key可能共享同一个锁，但比全局锁的并发度更高。

## 核心思想

```
传统全局锁:        所有key竞争同一个锁
┌─────────────────┐
│   Global Lock   │
└─────────────────┘
    key1,key2,key3... 全部互斥

分段锁:           key分散到不同锁
┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
│Lock0 │ │Lock1 │ │Lock2 │ │Lock3 │
└──────┘ └──────┘ └──────┘ └──────┘
key1    key2    key3    key4    key5...
```

## 工作原理

### 1. 初始化
```go
// 创建4个分段锁
segLock := NewSegmentKeysLock(4)

// 内部创建4个RWMutex
locks = [lock0, lock1, lock2, lock3]
```

### 2. Hash分配
```go
func (s *SegmentKeysLock) getLock(key string) *sync.RWMutex {
    hash := s.hash(key)        // 计算hash值，如 fnv("user_123") = 1234567890
    return s.locks[hash % s.size]  // 取模得到锁索引，如 1234567890 % 4 = 2
}
```

### 3. 锁操作
```go
// 对特定key加锁
segLock.Lock("user_123")   // 获取锁索引2的写锁
segLock.Unlock("user_123") // 释放锁索引2的写锁
```

## 完整示例

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func main() {
    // 创建16个分段的分段锁
    segLock := NewSegmentKeysLock(16)
    
    var wg sync.WaitGroup
    
    // 并发操作不同的key
    keys := []string{"user_1", "user_2", "user_3", "user_100"}
    
    for _, key := range keys {
        wg.Add(1)
        go func(k string) {
            defer wg.Done()
            
            // 写锁保护
            segLock.Lock(k)
            defer segLock.Unlock(k)
            
            // 模拟业务操作
            fmt.Printf("正在处理 %s\n", k)
            time.Sleep(100 * time.Millisecond)
        }(key)
    }
    
    wg.Wait()
}
```

## 锁分配示意

```go
// 假设size=4，hash取模结果
"user_1"   → hash=1234567890 %4 = 2 → 使用 locks[2]
"user_2"   → hash=9876543210 %4 = 2 → 使用 locks[2] (冲突)
"order_1"  → hash=1111111111 %4 = 3 → 使用 locks[3]
"product_1"→ hash=2222222222 %4 = 0 → 使用 locks[0]

// locks[2] 被 user_1 和 user_2 共享
// 其他锁独立
```

## 为什么分段锁更好？

### 全局锁（性能差）
```go
type GlobalLock struct {
    mu sync.RWMutex
    data map[string]interface{}
}

// 所有key操作都会互斥
func (g *GlobalLock) Get(key string) {
    g.mu.RLock()         // 读锁全局阻塞
    defer g.mu.RUnlock()
    return g.data[key]
}
```

### 分段锁（性能好）
```go
// 不同段的key可以并行操作
func (s *SegmentKeysLock) Get(key string) {
    lock := s.getLock(key)
    lock.RLock()          // 只锁这一段
    defer lock.RUnlock()
    // 其他段的key不受影响
}
```

## 读写锁特性

```go
// 读锁共享（多个goroutine可以同时读同一个key）
segLock.RLock("user_1")
segLock.RLock("user_1")  // 允许多个读锁
segLock.RUnlock("user_1")
segLock.RUnlock("user_1")

// 写锁互斥（写时不能读也不能写）
segLock.Lock("user_1")   // 获取写锁
// 此时其他goroutine的RLock/Lock都会阻塞
segLock.Unlock("user_1")

// 尝试锁（非阻塞）
if segLock.TryLock("user_1") {
    defer segLock.Unlock("user_1")
    // 获取锁成功
} else {
    // 锁被占用，执行其他逻辑
}
```

## 性能对比

```go
// 基准测试示例
func BenchmarkGlobalLock(b *testing.B) {
    mu := &sync.RWMutex{}
    for i := 0; i < b.N; i++ {
        mu.RLock()
        mu.RUnlock()
    }
}

func BenchmarkSegmentLock(b *testing.B) {
    segLock := NewSegmentKeysLock(64)
    for i := 0; i < b.N; i++ {
        key := fmt.Sprintf("key_%d", i%100)
        segLock.RLock(key)
        segLock.RUnlock(key)
    }
}

// 结果（示例）：
// BenchmarkGlobalLock-8    50000000    30.5 ns/op
// BenchmarkSegmentLock-8   50000000    28.2 ns/op  // 略快
// 高并发场景下差距更明显
```

## 实际应用场景

### 1. 缓存系统
```go
type Cache struct {
    segLock *SegmentKeysLock
    data    map[string]interface{}
}

func (c *Cache) Get(key string) interface{} {
    c.segLock.RLock(key)
    defer c.segLock.RUnlock(key)
    return c.data[key]
}

func (c *Cache) Set(key string, value interface{}) {
    c.segLock.Lock(key)
    defer c.segLock.Unlock(key)
    c.data[key] = value
}
```

### 2. 计数器
```go
type Counter struct {
    segLock *SegmentKeysLock
    counts  map[string]int64
}

func (c *Counter) Incr(key string) {
    c.segLock.Lock(key)
    defer c.segLock.Unlock(key)
    c.counts[key]++
}
```

### 3. Session管理
```go
type SessionManager struct {
    segLock *SegmentKeysLock
    sessions map[string]*Session
}

func (sm *SessionManager) Get(sessionID string) *Session {
    sm.segLock.RLock(sessionID)
    defer sm.segLock.RUnlock(sessionID)
    return sm.sessions[sessionID]
}
```

## 段数选择建议

```go
// 段数太少：锁竞争激烈
segLock := NewSegmentKeysLock(4)   // 只有4个锁，容易冲突

// 段数太多：内存浪费
segLock := NewSegmentKeysLock(1024) // 1024个锁对象

// 经验值：
// - 并发量小：16-32
// - 并发量大：64-256
// - 通常2的幂次方（便于取模优化）
```

## 线程安全分析

```go
// ✅ 安全：不同key并行操作
go segLock.Lock("user_1")
go segLock.Lock("user_2")  // 可能不同锁，并行执行

// ⚠️ 注意：相同key串行
go segLock.Lock("user_1")
go segLock.Lock("user_1")  // 必须等待第一个释放

// ⚠️ 注意：hash冲突导致不同key互斥
// "user_1" 和 "user_2" 可能hash到同一个锁
go segLock.Lock("user_1")
go segLock.Lock("user_2")  // 可能阻塞（如果hash冲突）
```

## 优缺点总结

**优点**：
- ✅ 减少锁竞争，提高并发度
- ✅ 支持读写锁分离
- ✅ 支持TryLock非阻塞尝试

**缺点**：
- ❌ 内存开销（多个锁对象）
- ❌ 可能hash冲突导致性能下降
- ❌ 无法保证跨多个key的原子性（需要全局锁）

## 与sync.Map对比

| 特性 | SegmentKeysLock + map | sync.Map |
|------|----------------------|----------|
| 适用场景 | 读写都频繁 | 读多写少 |
| 内存开销 | 较大 | 较小 |
| 类型安全 | 泛型可保证 | 需要类型断言 |
| 灵活性 | 高（手动控制） | 低（固定API） |

这是一个典型的**空间换时间**的优化方案！