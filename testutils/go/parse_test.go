package goutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeserialize(t *testing.T) {
	assert.NotPanics(
		t, func() {
			v1 := Deserialize[int]("123")
			assert.Equal(t, 123, v1)
		},
	)
	assert.NotPanics(
		t, func() {
			v2 := Deserialize[string](`"abc"`)
			assert.Equal(t, "abc", v2)
		},
	)
	assert.NotPanics(
		t, func() {
			v3 := Deserialize[byte](`'a'`)
			assert.Equal(t, byte('a'), v3)
		},
	)
	assert.NotPanics(
		t, func() {
			v4 := Deserialize[[]int]("[]")
			assert.Equal(t, []int{}, v4)
		},
	)
	assert.NotPanics(
		t, func() {
			v5 := Deserialize[[]int]("[1,2,3]")
			assert.Equal(t, []int{1, 2, 3}, v5)
		},
	)
	assert.NotPanics(
		t, func() {
			v6 := Deserialize[[]string](`["a","b","c"]`)
			assert.Equal(t, []string{"a", "b", "c"}, v6)
		},
	)
	assert.NotPanics(
		t, func() {
			v7 := Deserialize[*TreeNode]("[1,2,3]")
			assert.Equal(t, 1, v7.Val)
			assert.Equal(t, 2, v7.Left.Val)
			assert.Equal(t, 3, v7.Right.Val)
		},
	)
	assert.NotPanics(
		t, func() {
			v8 := Deserialize[*ListNode]("[1,2,3]")
			assert.Equal(t, 1, v8.Val)
			assert.Equal(t, 2, v8.Next.Val)
			assert.Equal(t, 3, v8.Next.Next.Val)
		},
	)
	assert.NotPanics(
		t, func() {
			v9 := Deserialize[float64]("1.2")
			assert.Equal(t, 1.2, v9)
		},
	)
	assert.NotPanics(
		t, func() {
			v10 := Deserialize[bool]("true")
			assert.Equal(t, true, v10)
		},
	)
	assert.NotPanics(
		t, func() {
			v11 := Deserialize[bool]("false")
			assert.Equal(t, false, v11)
		},
	)
	assert.NotPanics(
		t, func() {
			v12 := Deserialize[[][]int]("[[1,2],[3,4]]")
			assert.Equal(t, [][]int{{1, 2}, {3, 4}}, v12)
		},
	)
	assert.NotPanics(
		t, func() {
			v13 := Deserialize[[]*TreeNode]("[[1,2,3],[4,5,6]]")
			assert.Len(t, v13, 2)
		},
	)
	assert.Panics(t, func() { Deserialize[bool]("True") })
	assert.Panics(t, func() { Deserialize[func()]("1") })
	assert.Panics(t, func() { Deserialize[int](`"1.2"`) })
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSplitArray(t *testing.T) {
	tests := []struct {
		input string
		want  []string
		err   error
	}{
		{"[]", []string{}, nil},
		{"[1]", []string{"1"}, nil},
		{"[[1], [2]]", []string{"[1]", "[2]"}, nil},
		{"[1,2,3]", []string{"1", "2", "3"}, nil},
		{"[1, 2, 3]", []string{"1", "2", "3"}, nil},
		{" [1,2,3] ", []string{"1", "2", "3"}, nil},
		{`[1, "2, 3"]`, []string{"1", `"2, 3"`}, nil},
		{"[1,2,3,]", nil, fmt.Errorf("invalid array: [1,2,3,]")},   // trailing comma
		{"[1,2,3", nil, fmt.Errorf("invalid array: [1,2,3")},       // missing closing bracket
		{"1,2,3", nil, fmt.Errorf("invalid array: 1,2,3")},         // no brackets
		{`[1,2,"[","[]"]`, []string{"1", "2", `"["`, `"[]"`}, nil}, // string contains brackets
		{`[null,1,2,null]`, []string{"null", "1", "2", "null"}, nil},
	}

	for _, tc := range tests {
		got, err := SplitArray(tc.input)

		if !sliceEqual(tc.want, got) || (err != nil && tc.err != nil && err.Error() != tc.err.Error()) {
			t.Errorf("SplitArray(%q) = (%q, %v), want (%q, %v)", tc.input, got, err, tc.want, tc.err)
		}
	}
}
