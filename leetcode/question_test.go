package leetcode_test

import (
	"slices"
	"testing"

	"github.com/j178/leetgo/leetcode"
)

func TestGetFormattedContentHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "subsets keyword spacing",
			html: `<p>return <em>all possible</em> <span data-keyword="subset"><em>subsets</em></span> <em>(the power set)</em>.</p>`,
			want: "return *all possible* *subsets* *(the power set)*.\n",
		},
		{
			name: "valid parenthesis inline spacing",
			html: `<p>return <code>true</code> <em>if</em> <code>s</code> <em>is <strong>valid</strong></em>.</p>`,
			want: "return `true` *if* `s` *is **valid***.\n",
		},
		{
			name: "subscripts and superscripts in constraints and examples",
			html: "<p>a<sub>1</sub> &lt;= 10<sup>2</sup></p><p><code>a<sub>1</sub> &lt;= 10<sup>2</sup></code></p><pre><strong>Output:</strong> 10<sup>2</sup>\n</pre><p>Explanation</p>",
			want: "a₁ &lt;= 10²\n\n`a₁ <= 10²`\n\n```\nOutput: 10²\n```\n\nExplanation\n",
		},
		{
			name: "example indentation and trailing newline",
			html: "<pre><strong>Input:</strong> tree =\n  1\n / \\\n2   3\n</pre><p>Explanation</p>",
			want: "```\nInput: tree =\n  1\n / \\\n2   3\n```\n\nExplanation\n",
		},
		{
			name: "table with strikethrough and line breaks",
			html: "<table><thead><tr><th>State</th><th>Value</th></tr></thead><tbody><tr><td><del>old</del><br>new</td><td><code>0</code></td></tr></tbody></table>",
			want: "| State | Value |\n|---|---|\n| ~~old~~  <br />new | `0` |\n",
		},
		{
			name: "task list checkboxes",
			html: `<ul><li><input type="checkbox" checked disabled>Read the input</li><li><input type="checkbox" disabled> Return the result</li></ul>`,
			want: "- [x] Read the input\n- [ ] Return the result\n",
		},
		{
			name: "entities and invisible characters",
			html: `<p>所有&nbsp;<code>nums</code>&#8203; 的值都满足 <code>0 &lt;= nums[i]</code>。</p>`,
			want: "所有 `nums` 的值都满足 `0 <= nums[i]`。\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := leetcode.QuestionData{Content: tt.html, EditorType: leetcode.EditorTypeCKEditor}
			if got := q.GetFormattedContent(); got != tt.want {
				t.Errorf("GetFormattedContent() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestParseExampleOutputs(t *testing.T) {
	tests := []struct {
		name    string
		editor  leetcode.EditorType
		content string
		want    []string
	}{
		{
			name:    "output wrapped in code",
			editor:  leetcode.EditorTypeCKEditor,
			content: "<p><strong>Input:</strong> <code>nums = [2,7,11,15], target = 9</code></p>\n\n<p><strong>Output:</strong> <code>[0,1]</code></p>\n\n<p><strong>Explanation:</strong></p>\n\n<p>The sum of <code>nums[0]</code> and <code>nums[1]</code> is <code>2 + 7 = 9</code>, which equals the target. Therefore, the indices <code>[0, 1]</code> are returned.</p>\n\n<p><strong>Example 2:</strong></p>\n\n<p><strong>Input:</strong> <code>nums = [3,2,4], target = 6</code></p>\n\n<p><strong>Output:</strong> <code>[1,2]</code></p>",
			want:    []string{"[0,1]", "[1,2]"},
		},
		{
			name:    "plain output in pre",
			editor:  leetcode.EditorTypeCKEditor,
			content: "<pre>\n<strong>Input:</strong> nums = [2,7,11,15], target = 9\n<strong>Output:</strong> [0,1]\n<strong>Explanation:</strong> Because nums[0] + nums[1] == 9, we return [0, 1].\n</pre>",
			want:    []string{"[0,1]"},
		},
		{
			name:    "output wrapped in span",
			editor:  leetcode.EditorTypeCKEditor,
			content: "<p><strong>Input:</strong> <span class=\"example-io\">nums1 = [1,3,4], nums2 = [1,3,4], k = 1</span></p>\n\n<p><strong>Output:</strong> <span class=\"example-io\">5</span></p>",
			want:    []string{"5"},
		},
		{
			name:    "chinese output",
			editor:  leetcode.EditorTypeCKEditor,
			content: "<pre>\n<strong>输入：</strong>nums = [2,7,11,15], target = 9\n<strong>输出：</strong>[0,1]\n<strong>解释：</strong>因为 nums[0] + nums[1] == 9 ，返回 [0, 1] 。\n</pre>",
			want:    []string{"[0,1]"},
		},
		{
			name:    "markdown output wrapped in code",
			editor:  leetcode.EditorTypeMarkdown,
			content: "Input: nums = [2,7,11,15], target = 9\nOutput: <code>[0,1]</code>\n",
			want:    []string{"[0,1]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := leetcode.QuestionData{Content: tt.content, EditorType: tt.editor}
			got := q.ParseExampleOutputs()
			if !slices.Equal(got, tt.want) {
				t.Errorf("ParseExampleOutputs() = %q; want %q", got, tt.want)
			}
		})
	}
}
