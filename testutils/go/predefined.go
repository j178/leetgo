package goutils

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

var ErrInfiniteLoop = errors.New("infinite loop detected")

/*
Much appreciated to EndlessCheng
Adapted from https://github.com/EndlessCheng/codeforces-go/blob/ae5b312f3f/leetcode/testutil/leetcode.go
*/

type ListNode struct {
	Val  int
	Next *ListNode
}

func DeserializeListNode(s string) (*ListNode, error) {
	var res []*int
	err := json.Unmarshal([]byte(s), &res)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	root := &ListNode{}
	n := root
	for i := 0; i < len(res)-1; i++ {
		n.Val = *res[i]
		n.Next = &ListNode{}
		n = n.Next
	}
	n.Val = *res[len(res)-1]
	return root, nil
}

func (l *ListNode) ToString() string {
	seen := make(map[*ListNode]bool, 10)

	sb := &strings.Builder{}
	sb.WriteByte('[')
	for ; l != nil; l = l.Next {
		if sb.Len() > 1 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Itoa(l.Val))

		if seen[l] {
			panic(ErrInfiniteLoop)
		}
		seen[l] = true
	}
	sb.WriteByte(']')
	return sb.String()
}

func (l *ListNode) Values() []int {
	vals := []int{}
	for ; l != nil; l = l.Next {
		vals = append(vals, l.Val)
	}
	return vals
}

func (l *ListNode) Nodes() []*ListNode {
	nodes := []*ListNode{}
	for ; l != nil; l = l.Next {
		nodes = append(nodes, l)
	}
	return nodes
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func DeserializeTreeNode(s string) (*TreeNode, error) {
	var res []*int
	err := json.Unmarshal([]byte(s), &res)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	nodes := make([]*TreeNode, len(res))
	for i := 0; i < len(res); i++ {
		if res[i] != nil {
			nodes[i] = &TreeNode{Val: *res[i]}
		}
	}
	root := nodes[0]
	for i, j := 0, 1; j < len(res); i++ {
		if nodes[i] != nil {
			nodes[i].Left = nodes[j]
			j++
			if j >= len(res) {
				break
			}
			nodes[i].Right = nodes[j]
			j++
			if j >= len(res) {
				break
			}
		}
	}
	return root, nil
}

func (t *TreeNode) ToString() string {
	nodes := []*TreeNode{}
	queue := []*TreeNode{t}
	seen := make(map[*TreeNode]bool, 10)
	for len(queue) > 0 {
		t, queue = queue[0], queue[1:]
		nodes = append(nodes, t)
		if t != nil {
			if seen[t] {
				panic(ErrInfiniteLoop)
			}
			seen[t] = true

			queue = append(queue, t.Left, t.Right)
		}
	}

	for len(nodes) > 0 && nodes[len(nodes)-1] == nil {
		nodes = nodes[:len(nodes)-1]
	}

	sb := &strings.Builder{}
	sb.WriteByte('[')
	for _, node := range nodes {
		if sb.Len() > 1 {
			sb.WriteByte(',')
		}
		if node != nil {
			sb.WriteString(strconv.Itoa(node.Val))
		} else {
			sb.WriteString("null")
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

type NaryTreeNode struct {
	Val      int
	Children []*NaryTreeNode
}

func DeserializeNaryTreeNode(s string) (*NaryTreeNode, error) {
	var res []*int
	if err := json.Unmarshal([]byte(s), &res); err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	// 用一个伪的头结点
	root := &NaryTreeNode{}
	q := []*NaryTreeNode{root}
	for i := 0; i < len(res); i++ {
		node := q[0]
		q = q[1:]
		for ; i < len(res) && res[i] != nil; i++ {
			n := &NaryTreeNode{Val: *res[i]}
			node.Children = append(node.Children, n)
			q = append(q, n)
		}
	}

	return root.Children[0], nil
}

func (t *NaryTreeNode) ToString() string {
	nodes := []*NaryTreeNode{}
	q := []*NaryTreeNode{{Children: []*NaryTreeNode{t}}}

	for len(q) > 0 {
		node := q[0]
		q = q[1:]
		nodes = append(nodes, node)

		if node != nil {
			if len(node.Children) > 0 {
				q = append(q, node.Children...)
			}
			q = append(q, nil)
		}
	}
	// 去除头结点
	nodes = nodes[1:]
	// 去除末尾的 null
	for len(nodes) > 0 && nodes[len(nodes)-1] == nil {
		nodes = nodes[:len(nodes)-1]
	}

	sb := strings.Builder{}
	sb.WriteByte('[')
	for _, node := range nodes {
		if sb.Len() > 1 {
			sb.WriteByte(',')
		}
		if node == nil {
			sb.WriteString("null")
		} else {
			sb.WriteString(strconv.Itoa(node.Val))
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

type Node struct {
	Val   int
	Prev  *Node
	Next  *Node
	Child *Node
}
