package runes

import (
	"testing"
)

func TestIndex(t *testing.T) {
	str := "你好啊今天天气怎么样！"
	needle := []rune("今天")
	index := Index([]rune(str), needle)
	t.Logf("index: %d", index)
}
