package github

import (
	"fmt"
	"strconv"
)

// ptr 返回值指针的泛型辅助函数
func ptr[T any](v T) *T { return &v }

// itoa 将 int 转换为字符串（用于构建 URL 路径）
func itoa(n int) string { return strconv.Itoa(n) }

// fmtInt64 将 int64 转换为字符串
func fmtInt64(n int64) string { return fmt.Sprint(n) }
