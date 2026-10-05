package lang

import (
	"testing"

	"github.com/j178/leetgo/leetcode"
)

func TestToRustVarName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"query", "query"},
		{"isValidBST", "is_valid_bst"},
	}
	for _, tt := range tests {
		if got := toRustVarName(tt.name); got != tt.want {
			t.Errorf("toRustVarName(%v) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestGenerateRustNormalTestCode(t *testing.T) {
	cases := []struct {
		name       string
		params     []leetcode.MetaDataParam
		returnType string
		output     *leetcode.MetaDataOutput
		code       string
		want       string
	}{
		{
			name:       "sum",
			params:     []leetcode.MetaDataParam{{Name: "nums", Type: "integer[]"}},
			returnType: "integer",
			code:       "impl Solution { pub fn sum(mut nums: Vec<i32>) -> i32 {} }",
			want: `fn main() -> Result<()> {
	let nums = deserialize::<Vec<i32>>(&read_line()?)?;
	let ans: i32 = Solution::sum(nums);

	println!("\noutput: {}", serialize(ans)?);
	Ok(())
}`,
		},
		{
			name: "removeElement",
			params: []leetcode.MetaDataParam{
				{Name: "nums", Type: "integer[]"},
				{Name: "val", Type: "integer"},
			},
			returnType: "integer",
			code:       "impl Solution { pub fn remove_element(nums: &mut Vec<i32>, val: i32) -> i32 {} }",
			want: `fn main() -> Result<()> {
	let mut nums = deserialize::<Vec<i32>>(&read_line()?)?;
	let val = deserialize::<i32>(&read_line()?)?;
	let ans: i32 = Solution::remove_element(&mut nums, val);

	println!("\noutput: {}", serialize(ans)?);
	Ok(())
}`,
		},
		{
			name: "merge",
			params: []leetcode.MetaDataParam{
				{Name: "nums1", Type: "integer[]"},
				{Name: "m", Type: "integer"},
				{Name: "nums2", Type: "integer[]"},
				{Name: "n", Type: "integer"},
			},
			returnType: "void",
			output:     &leetcode.MetaDataOutput{ParamIndex: 0},
			code: `impl Solution {
    pub fn merge(
        nums1: &mut Vec<i32>, m: i32,
        nums2: &mut Vec<i32>, n: i32,
    ) {}
}`,
			want: `fn main() -> Result<()> {
	let mut nums1 = deserialize::<Vec<i32>>(&read_line()?)?;
	let m = deserialize::<i32>(&read_line()?)?;
	let mut nums2 = deserialize::<Vec<i32>>(&read_line()?)?;
	let n = deserialize::<i32>(&read_line()?)?;
	Solution::merge(&mut nums1, m, &mut nums2, n);
	let ans: Vec<i32> = nums1;

	println!("\noutput: {}", serialize(ans)?);
	Ok(())
}`,
		},
		{
			name:       "flatten",
			params:     []leetcode.MetaDataParam{{Name: "root", Type: "TreeNode"}},
			returnType: "void",
			output:     &leetcode.MetaDataOutput{ParamIndex: 0},
			code:       "impl Solution { pub fn flatten(root: &mut Option<Rc<RefCell<TreeNode>>>) {} }",
			want: `fn main() -> Result<()> {
	let mut root = deserialize::<BinaryTree>(&read_line()?)?.into();
	Solution::flatten(&mut root);
	let ans: BinaryTree = root.into();

	println!("\noutput: {}", serialize(ans)?);
	Ok(())
}`,
		},
		{
			name:       "reorderList",
			params:     []leetcode.MetaDataParam{{Name: "head", Type: "ListNode"}},
			returnType: "void",
			output:     &leetcode.MetaDataOutput{ParamIndex: 0},
			code:       "impl Solution { pub fn reorder_list(head: &mut Option<Box<ListNode>>) {} }",
			want: `fn main() -> Result<()> {
	let mut head = deserialize::<LinkedList>(&read_line()?)?.into();
	Solution::reorder_list(&mut head);
	let ans: LinkedList = head.into();

	println!("\noutput: {}", serialize(ans)?);
	Ok(())
}`,
		},
		{
			name:       "reverseList",
			params:     []leetcode.MetaDataParam{{Name: "head", Type: "ListNode"}},
			returnType: "ListNode",
			code:       "impl Solution { pub fn reverse_list(mut head: Option<Box<ListNode>>) -> Option<Box<ListNode>> {} }",
			want: `fn main() -> Result<()> {
	let head = deserialize::<LinkedList>(&read_line()?)?.into();
	let ans: LinkedList = Solution::reverse_list(head).into();

	println!("\noutput: {}", serialize(ans)?);
	Ok(())
}`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			q := &leetcode.QuestionData{
				MetaData: leetcode.MetaData{
					Name:   tt.name,
					Params: tt.params,
					Return: &leetcode.MetaDataReturn{Type: tt.returnType},
					Output: tt.output,
				},
				CodeSnippets: []leetcode.CodeSnippet{{LangSlug: "rust", Code: tt.code}},
			}
			mainCode, err := (rust{}).generateNormalTestCode(q)
			if err != nil {
				t.Fatal(err)
			}
			assertSnapshot(t, mainCode, tt.want)
		})
	}
}
