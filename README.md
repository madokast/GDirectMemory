# 直接内存

direct 包提供不受 Go GC 内存管理的直接内存访问，并提供对应的集合类型。

## 内存初始化

direct 包含一个全局内存对象，使用如下方法申请一个 10 MB 的内存空间。

根据操作系统机制，申请内存仅仅是 commit，并未实际物理空间。

当内存使用结束时，使用 Free() 释放申请的空间。

```go
direct.Global.Init(10 * direct.MB)
defer direct.Global.Free()
```

## 手动释放

从 direct 中申请的内存对象，不受 Go GC 管控，需要手动释放，否则将导致内存泄漏。

memory/trace.go 提供了简单的内存追踪器，可以提供内存泄漏信息。

因为内存追踪对性能很大，默认关闭，在 trace.go 中修改源码 `const Trace = true` 开启。

开启后，可以通过 `direct.Global.MemoryLeakInfo()` 打印泄露信息。

```go
func TestSliceLeak(t *testing.T) {
    direct.Global.Init(10 * direct.MB)
    defer direct.Global.Free()
    
    var s, _ = direct.MakeSliceFromGoSlice([]int{1, 2, 3}) // line:35
    
    // [1 2 3]
    fmt.Println(s)
    
    //s.Free()
    
    // Addr:0x1FCC64A0340 index:4 type:Slice size:256B allocated at D:/learn/repo/go_repo/direct/example/slice_example_test.go:35
    fmt.Println(direct.Global.MemoryLeakInfo())
}
```
下面的代码创建了 Slice 对象，但是没有释放内存 `s.Free()`，打印内存泄露信息。

`Addr:0x1FCC64A0340 index:4 type:Slice size:256B allocated at D:/learn/repo/go_repo/direct/example/slice_example_test.go:35`

可以看到泄露的内存对象为 Slice 类型，申请该内存的地址为 `slice_example_test.go:35`，其中 35 表示这个文件的第 35 行。

## 捕获 OOM 错误

和 Go 的内存管理不同，当内存不足时，申请内存时将返回 OOM 错误，可以捕获并处理内存不足错误。

```go
func TestSliceOOM(t *testing.T) {
	direct.Global.Init(10 * direct.KB)
	defer direct.Global.Free()

	var s, err = direct.MakeSlice[int](1024 * 1024)
	defer s.Free()

	if _, ok := err.(*direct.OOMError); ok {
		fmt.Println("OOM", err) // OOM out of memory when alloc 32769 pages. The memory details is ...
	}
}
```

## 集合类型

direct 提供了 Slice，Map 两个常用集合类型。

使用方法和 Java 集合对象类型，详见代码方法定义。

## 字符串

direct 中使用字符串工厂 `StringFactory` 构建字符串对象。示例代码如下

```go
func TestStringFactory(t *testing.T) {
	Global.Init(1 * MB)
	defer Global.Free()

	factory := NewStringFactory()
	defer factory.Destroy()

	s1, err := factory.CreateFromGoString("hello")
	utils.PanicErr(err)
	defer s1.Free()

	t.Log(s1.Length())         // 5
	t.Log(s1)                  // hello
	t.Log(s1.AsGoString())     // hello
	t.Log(s1.CopyToGoString()) // hello

	utils.Assert(s1.CopyToGoString() == "hello")
}
```

注意方法 AsGoString() 仅仅将 direct 字符串对象转为 Go 字符串对象，二者生命周期相同（共用相同的地址），CopyToGoString() 将 direct 字符串赋值为 Go 字符串。

## MOVE 移动机制

类型 C++ 中的 move，direct 集合对象可以移动，被移动的内存对象释放时，不会造成 double free 错误。

下面代码将 strings 切片中的一部分元素 move 到了 targetStrings 中。

最终释放所有元素内存时，不用关心元素是否被移动。

```go
func TestRefCnt(t *testing.T) {
	rand.Seed(1)
	direct.Global.Init(1 * direct.MB)
	defer direct.Global.Free()

	var stringFactory = direct.NewStringFactory()
	defer func() { stringFactory.Destroy() }()

	var strings direct.Slice[direct.String] // []string
	defer func() { strings.Free() }()

	// 存放 10 个字符串
	for i := 0; i < 10; i++ {
		s, _ := stringFactory.CreateFromGoString(strconv.Itoa(rand.Intn(1000)))
		_ = strings.Append(s)
	}
	fmt.Println(strings) // [81 887 847 59 81 318 425 540 456 300]

	var targetStrings direct.Slice[direct.String]
	defer func() { targetStrings.Free() }()

	// 筛选字符串
	for i := direct.SizeType(0); i < strings.Length(); i++ {
		if strings.Get(i).Length()%2 == 0 {
			_ = targetStrings.Append(strings.RefAt(i).Move()) // move
		}
	}
	fmt.Println(strings)       // [ 887 847   318 425 540 456 300]
	fmt.Println(targetStrings) // [81 59 81]

	// 释放两个字符串数组的元素
	strings.Iterate(func(s direct.String) { s.Free() })
	targetStrings.Iterate(func(s direct.String) { s.Free() })
}
```

## 引用计数

手动内存管理，常需要引用计数来释放共享内存对象。简答代码如下

```go
func TestRefCnt(t *testing.T) {
	rand.Seed(1)
	direct.Global.Init(1 * direct.MB)
	defer direct.Global.Free()

	var refCntFactory = direct.CreateSharedFactory[direct.Slice[int]]()
	defer func() { refCntFactory.Destroy() }()

	var ss direct.Slice[direct.Shared[direct.Slice[int]]] // [][]int
	defer func() { ss.Free() }()

	// 存放 10 个切片，其中奇数位置使用前一个元素的引用
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			var s, _ = direct.MakeSliceFromGoSlice([]int{i}) // 创建切片
			shared, _ := refCntFactory.MakeShared(s)         // 切片转为引用
			_ = ss.Append(shared)
		} else {
			pre := ss.RefAt(direct.SizeType(i - 1))
			_ = ss.Append(pre.Share()) // share
		}
	}
	fmt.Println(ss) // [[0] [0] [2] [2] [4] [4] [6] [6] [8] [8]]

	var targetStrings direct.Slice[direct.String]
	defer func() { targetStrings.Free() }()

	// 释放元素
	ss.Iterate(func(s direct.Shared[direct.Slice[int]]) { s.Free() })
}
```