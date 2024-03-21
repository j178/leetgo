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

	naryTree := &NaryTreeNode{Val: 1}
	naryTree.Children = []*NaryTreeNode{{Val: 2, Children: []*NaryTreeNode{naryTree}}}

	type toStringer interface {
		ToString() string
	}

	tests := []toStringer{
		linkedList,
		tree,
		naryTree,
	}

	for _, tc := range tests {
		assert.PanicsWithValue(t, ErrInfiniteLoop, func() { tc.ToString() })
	}
}
