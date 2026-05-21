## 2. Pair 包解释

这是一个**泛型键值对工具包**，提供了多种数据转换方法。

### 核心结构
```go
type Pair[K any, V any] struct {
    Key   K  // 键，任意类型
    Value V  // 值，任意类型
}
```

### 各函数作用

#### `NewPair` - 创建单个键值对
```go
pair := NewPair("name", "张三")  // Pair[string, string]
```

#### `NewPairs` - 批量创建（键数组+值数组 → Pair数组）
```go
keys := []string{"a", "b", "c"}
values := []int{1, 2, 3}
pairs, _ := NewPairs(keys, values)
// 结果: [{a,1}, {b,2}, {c,3}]
```

#### `SplitPairs` - 反向拆分（Pair数组 → 键数组+值数组）
```go
pairs := []Pair[string, int]{{"a",1}, {"b",2}}
keys, values := SplitPairs(pairs)
// keys: ["a","b"], values: [1,2]
```

#### `FlattenPairs` - 扁平化（Pair数组 → 交替数组）
```go
pairs := []Pair[string, int]{{"a",1}, {"b",2}}
flat := FlattenPairs(pairs)
// flat: ["a", 1, "b", 2]  // 注意类型变成 []any
```

#### `PackPairs` - 打包（交替数组 → Pair数组）
```go
flat := []any{"a", 1, "b", 2}
pairs := PackPairs[string, int](flat)
// pairs: [{a,1}, {b,2}]
```
⚠️ **注意**：这里用了类型断言 `.(K)` 和 `.(V)`，类型不匹配会 panic

### 实际应用场景

在 `randx` 包中的使用：
```go
var typeCharsetPairs = []pair.Pair[Type, string]{
    pair.NewPair(TypeDigit, "0123456789"),
    pair.NewPair(TypeLowerCase, "abcdefghijklmnopqrstuvwxyz"),
    // ...
}

// 遍历判断
for _, p := range typeCharsetPairs {
    if (typ & p.Key) == p.Key {
        charset += p.Value
    }
}
```

这种设计的好处：
1. **类型安全**：Key 固定为 Type，Value 固定为 string
2. **可扩展**：方便添加新字符类型
3. **易维护**：数据和逻辑分离

### 使用示例对比

**传统方式（不用Pair）**：
```go
charsetMap := map[Type]string{
    TypeDigit: "0123456789",
    TypeLowerCase: "abc...",
}
```

**使用Pair的优势**：
- 保持顺序（map 无序）
- 方便遍历和转换
- 提供丰富的工具函数（SplitPairs、FlattenPairs等）