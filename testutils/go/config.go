package common

import "time"

var (
	// 用于编写 struct + method 的题目，方便打断点
	DebugCallIndex int

	DebugTLE = 2 * time.Second

	AssertOutput = true
)
