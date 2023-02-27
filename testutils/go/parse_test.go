package goutils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeserialize(t *testing.T) {
	v1, err := Deserialize[int]("123")
	assert.NoError(t, err)
	assert.Equal(t, 123, v1)

	v2, err := Deserialize[string](`"abc"`)
	assert.NoError(t, err)
	assert.Equal(t, "abc", v2)

	v3, err := Deserialize[byte](`'a'`)
	assert.NoError(t, err)
	assert.Equal(t, byte('a'), v3)

	v4, err := Deserialize[[]int]("[]")
	assert.NoError(t, err)
	assert.Equal(t, []int{}, v4)

	v5, err := Deserialize[[]int]("[1,2,3]")
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, v5)

	v6, err := Deserialize[[]string](`["a","b","c"]`)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, v6)

	v7, err := Deserialize[*TreeNode]("[1,2,3]")
	assert.NoError(t, err)
	assert.Equal(t, v7.Val, 1)
	assert.Equal(t, v7.Left.Val, 2)
	assert.Equal(t, v7.Right.Val, 3)

	v8, err := Deserialize[*ListNode]("[1,2,3]")
	assert.NoError(t, err)
	assert.Equal(t, v8.Val, 1)
	assert.Equal(t, v8.Next.Val, 2)
	assert.Equal(t, v8.Next.Next.Val, 3)

	v9, err := Deserialize[float64]("1.2")
	assert.NoError(t, err)
	assert.Equal(t, 1.2, v9)

	v10, err := Deserialize[bool]("true")
	assert.NoError(t, err)
	assert.Equal(t, true, v10)

	v11, err := Deserialize[bool]("false")
	assert.NoError(t, err)
	assert.Equal(t, false, v11)

	v12, err := Deserialize[[][]int]("[[1,2],[3,4]]")
	assert.NoError(t, err)
	assert.Equal(t, [][]int{{1, 2}, {3, 4}}, v12)

	v13, err := Deserialize[[]*TreeNode]("[[1,2,3],[4,5,6]]")
	assert.NoError(t, err)
	assert.Len(t, v13, 2)

	_, err = Deserialize[bool]("True")
	assert.Error(t, err)

	_, err = Deserialize[func()]("")
	assert.Error(t, err)

	_, err = Deserialize[int](`"1.2"`)
	assert.Error(t, err)
}

func TestDeserializeByGoType(t *testing.T) {
	v1, err := DeserializeByGoType("int", "123")
	assert.NoError(t, err)
	assert.EqualValues(t, 123, v1.Int())

	v2, err := DeserializeByGoType("string", `"abc"`)
	assert.NoError(t, err)
	assert.Equal(t, "abc", v2.String())

	v3, err := DeserializeByGoType("byte", `'a'`)
	assert.NoError(t, err)
	assert.Equal(t, uint8('a'), uint8(v3.Uint()))

	v4, err := DeserializeByGoType("[]int", "[]")
	assert.NoError(t, err)
	assert.Equal(t, reflect.Slice, v4.Type().Kind())
	assert.Equal(t, reflect.Int, v4.Type().Elem().Kind())
	var v4Slice []int
	reflect.ValueOf(&v4Slice).Elem().Set(v4)
	assert.Equal(t, []int{}, v4Slice)

	v5, err := DeserializeByGoType("[]int", "[1,2,3]")
	assert.NoError(t, err)
	assert.Equal(t, reflect.Slice, v4.Type().Kind())
	assert.Equal(t, reflect.Int, v4.Type().Elem().Kind())
	var v5Slice []int
	reflect.ValueOf(&v5Slice).Elem().Set(v5)
	assert.Equal(t, []int{1, 2, 3}, v5Slice)

	v6, err := DeserializeByGoType("[]string", `["a","b","c"]`)
	assert.NoError(t, err)
	var v6Slice []string
	reflect.ValueOf(&v6Slice).Elem().Set(v6)
	assert.Equal(t, []string{"a", "b", "c"}, v6Slice)

	v7, err := DeserializeByGoType("*TreeNode", "[1,2,3]")
	assert.NoError(t, err)
	var v7Node *TreeNode
	reflect.ValueOf(&v7Node).Elem().Set(v7)
	assert.Equal(t, v7Node.Val, 1)
	assert.Equal(t, v7Node.Left.Val, 2)
	assert.Equal(t, v7Node.Right.Val, 3)

	v8, err := DeserializeByGoType("*ListNode", "[1,2,3]")
	assert.NoError(t, err)
	var v8Node *ListNode
	reflect.ValueOf(&v8Node).Elem().Set(v8)
	assert.Equal(t, v8Node.Val, 1)
	assert.Equal(t, v8Node.Next.Val, 2)
	assert.Equal(t, v8Node.Next.Next.Val, 3)

	v9, err := DeserializeByGoType("float64", "1.2")
	assert.NoError(t, err)
	assert.Equal(t, float64(1.2), v9.Float())

	v10, err := DeserializeByGoType("bool", "true")
	assert.NoError(t, err)
	assert.Equal(t, true, v10.Bool())

	v11, err := DeserializeByGoType("bool", "false")
	assert.NoError(t, err)
	assert.Equal(t, false, v11.Bool())

	v12, err := DeserializeByGoType("[][]int", "[[1,2],[3,4]]")
	assert.NoError(t, err)
	var v12Slice [][]int
	reflect.ValueOf(&v12Slice).Elem().Set(v12)
	assert.Equal(t, [][]int{{1, 2}, {3, 4}}, v12Slice)

	v13, err := DeserializeByGoType("[]*TreeNode", "[[1,2,3],[4,5,6]]")
	assert.NoError(t, err)
	assert.Equal(t, v13.Len(), 2)

	_, err = DeserializeByGoType("bool", "True")
	assert.Error(t, err)

	_, err = DeserializeByGoType("func()", "")
	assert.Error(t, err)

	_, err = DeserializeByGoType("int", `"1.2"`)
	assert.Error(t, err)
}
