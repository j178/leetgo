package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfiniteLoopDetect(t *testing.T) {

	linkedList := &ListNode{Val: 1}
	linkedList.Next = &ListNode{Val: 2, Next: linkedList}

	tree := &TreeNode{Val: 1}
	tree.Left = &TreeNode{Val: 2, Right: tree}

	type toStringer interface {
		ToString() string
	}

	tests := []toStringer{
		linkedList,
		tree,
	}

	for _, tc := range tests {
		assert.PanicsWithValue(t, ErrInfiniteLoop, func() { tc.ToString() })
	}
}
