package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NaryTreeNodeToString(t *testing.T) {
	type testcase struct {
		tree     *NaryTreeNode
		expected string
	}
	tests := []testcase{
		{
			tree:     &NaryTreeNode{Val: 1},
			expected: "[1]",
		},
		{
			tree: &NaryTreeNode{
				Val: 1,
				Children: []*NaryTreeNode{
					{
						Val: 3,
						Children: []*NaryTreeNode{
							{Val: 5},
							{Val: 6},
						},
					},
					{Val: 2},
					{Val: 4},
				},
			},
			expected: "[1,null,3,2,4,null,5,6]",
		},
		{
			tree: &NaryTreeNode{
				Val: 1,
				Children: []*NaryTreeNode{
					{Val: 2},
					{
						Val: 3,
						Children: []*NaryTreeNode{
							{Val: 6},
							{
								Val: 7,
								Children: []*NaryTreeNode{
									{
										Val:      11,
										Children: []*NaryTreeNode{{Val: 14}},
									},
								},
							},
						},
					},
					{
						Val: 4,
						Children: []*NaryTreeNode{
							{Val: 8, Children: []*NaryTreeNode{{Val: 12}}},
						},
					},
					{
						Val: 5,
						Children: []*NaryTreeNode{
							{
								Val:      9,
								Children: []*NaryTreeNode{{Val: 13}},
							},
							{Val: 10},
						},
					},
				},
			},
			expected: "[1,null,2,3,4,5,null,null,6,7,null,8,null,9,10,null,null,11,null,12,null,13,null,null,14]",
		},
	}
	for _, test := range tests {
		t.Run(
			"", func(t *testing.T) {
				assert.Equal(t, test.expected, test.tree.ToString())
			},
		)
	}
}

func Test_DeserializeNaryTree(t *testing.T) {
	testcases := []string{
		"[1]",
		"[1,null,3,2,4,null,5,6]",
		"[1,null,2,3,4,5,null,null,6,7,null,8,null,9,10,null,null,11,null,12,null,13,null,null,14]",
	}
	for _, test := range testcases {
		t.Run(
			"", func(t *testing.T) {
				tree, err := DeserializeNaryTreeNode(test)
				if assert.NoError(t, err) {
					assert.Equal(t, test, tree.ToString())
				}
			},
		)
	}
}
