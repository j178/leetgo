package goutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerialize(t *testing.T) {

	cycleCaseHead := &ListNode{Val: 1}
	cycleCaseHead.Next = &ListNode{Val: 2, Next: cycleCaseHead}

	tests := []struct {
		input *ListNode
		want  string
	}{
		{&ListNode{Val: 1, Next: &ListNode{Val: 2, Next: &ListNode{Val: 3}}}, "[1,2,3]"},
		{cycleCaseHead, "[1,2,1..inf]"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.want, tc.input.ToString())
	}
}
