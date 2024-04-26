package example

import (
	"fmt"
	"github.com/madokast/direct"
	"math/rand"
	"strconv"
	"testing"
)

func TestMove(t *testing.T) {
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
			_ = targetStrings.Append(strings.RefAt(i).Move())
		}
	}
	fmt.Println(strings)       // [ 887 847   318 425 540 456 300]
	fmt.Println(targetStrings) // [81 59 81]

	// 释放两个字符串数组的元素
	strings.Iterate(func(s direct.String) { s.Free() })
	targetStrings.Iterate(func(s direct.String) { s.Free() })
}

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
