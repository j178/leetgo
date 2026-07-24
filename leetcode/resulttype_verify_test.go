package leetcode

import (
	"encoding/json"
	"testing"
)

// TestResultType 基于真实 LeetCode API 返回的 metadata 类型，验证 ResultType() 的正确性。
//
// 	5种真实场景：
//
//	场景 A：正常题目，Return 有值          → 直接返回 Return.Type
//	场景 B：System Design，Return=void+Output → 通过 Output.ParamIndex 取参数类型
//	场景 C：manual 题目，return 字段缺失     → 💥 root cause, 修复后返回 ""
//	场景 D：manual 题目，Return=void 且无 Output → 与 C 同理, 修复后返回 ""
//	场景 E：manual 题目，Return 有正常值     → 与 A 路径相同
func TestResultType(t *testing.T) {
	tests := []struct {
		name     string
		meta     MetaData
		expected string
	}{
		// ── 场景 A：正常题目 ──────────────────────────────────────────
		{
			name: "A: 正常题目 (twoSum)",
			meta: MetaData{
				Name: "twoSum",
				Params: []MetaDataParam{
					{Name: "nums", Type: "integer[]"},
					{Name: "target", Type: "integer"},
				},
				Return: &MetaDataReturn{Type: "integer[]"},
			},
			expected: "integer[]",
		},

		// ── 场景 B：System Design 题目 ────────────────────────────────
		{
			name: "B: System Design (ExamRoom #855)",
			meta: MetaData{
				Name:         "ExamRoom",
				Params:       []MetaDataParam{{Name: "inputs", Type: "integer[]"}, {Name: "inputs", Type: "integer[]"}},
				Return:       &MetaDataReturn{Type: "void"},
				Output:       &MetaDataOutput{ParamIndex: 1},
				SystemDesign: true,
				ClassName:    "ExamRoom",
				Constructor:  MetaDataConstructor{Params: []MetaDataParam{{Name: "n", Type: "integer"}}},
				Methods:      []MetaDataMethod{{Name: "seat", Return: MetaDataReturn{Type: "integer"}}},
			},
			expected: "integer[]",
		},

		// ── 场景 C：manual 题目，return 字段缺失 ──────────────────────
		// 真实代表：#470 implement-rand10-using-rand7
		// API 返回: {"name":"rand10","params":[{"name":"n","type":"integer"}],"manual":true}
		// 注意：没有 "return" key → Go 反序列化后 Return=nil, Output=nil
		{
			name: "C: return 缺失 (#470 rand10)",
			meta: MetaData{
				Name: "rand10",
				Params: []MetaDataParam{
					{Name: "n", Type: "integer"},
				},
				Return: nil, // JSON 中不存在该 key
				Output: nil, // JSON 中不存在该 key
				Manual: true,
			},
			expected: "",
		},

		// ── 场景 D：manual 题目，Return=void 且无 Output ─────────────────
		// 真实代表：#843 guess-the-word, #489 robot-room-cleaner
		// API 返回: {"name":"findSecretWord",...,"return":{"type":"void"},"manual":true}
		// 注意：没有 "output" key → Output=nil
		{
			name: "D: return=void 无 output (#843 guess-the-word)",
			meta: MetaData{
				Name: "findSecretWord",
				Params: []MetaDataParam{
					{Name: "secret", Type: "string"},
					{Name: "words", Type: "string[]"},
					{Name: "allowedGuesses", Type: "integer"},
				},
				Return: &MetaDataReturn{Type: "void"},
				Output: nil, // JSON 中不存在该 key
				Manual: true,
			},
			expected: "",
		},

		// ── 场景 E：manual 题目，Return 有正常值 ──────────────────────
		// 真实代表：#278 first-bad-version, #374 guess-number-higher-or-lower
		// API 返回: {"name":"firstBadVersion",...,"return":{"type":"integer"},"manual":true}
		{
			name: "E: manual 题目 return 正常 (#278 first-bad-version)",
			meta: MetaData{
				Name: "firstBadVersion",
				Params: []MetaDataParam{
					{Name: "n", Type: "integer"},
				},
				Return: &MetaDataReturn{Type: "integer"},
				Manual: true,
			},
			expected: "integer",
		},

		// ── 边界：System Design 中 Output 有效的 void 路径 ──────────
		// 验证 Output.ParamIndex 路径不被修复影响
		{
			name: "B-2: return=void, output 有效",
			meta: MetaData{
				Name:   "voidWithOutput",
				Params: []MetaDataParam{{Name: "a", Type: "integer"}, {Name: "result", Type: "string[]"}},
				Return: &MetaDataReturn{Type: "void"},
				Output: &MetaDataOutput{ParamIndex: 1},
			},
			expected: "string[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.meta.normalize()
			got := tt.meta.ResultType()
			if got != tt.expected {
				t.Errorf("ResultType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestResultTypeFromJSON 验证从原始 JSON 字符串反序列化后的完整链路。
// 模拟 UnmarshalJSON → normalize → ResultType 的真实数据流。
func TestResultTypeFromJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string // 一次转义的 JSON 字符串，模拟 GraphQL 返回的 metaData 字段
		expected string
	}{
		{
			// 场景 C 的真实 JSON（来自 leetcode.cn API 返回）
			// {"name":"rand10","params":[{"name":"n","type":"integer"}],"manual":true}
			name:     "#470 rand10 (return 缺失)",
			jsonStr:  `{"name":"rand10","params":[{"name":"n","type":"integer"}],"manual":true}`,
			expected: "",
		},
		{
			// 场景 D 的真实 JSON（来自 leetcode.cn API 返回）
			// {"name":"findSecretWord",...,"return":{"type":"void"},"manual":true}
			name:     "#843 guess-the-word (return=void, 无 output)",
			jsonStr:  `{"name":"findSecretWord","params":[{"name":"secret","type":"string"},{"name":"words","type":"string[]"},{"name":"allowedGuesses","type":"integer"}],"return":{"type":"void"},"manual":true}`,
			expected: "",
		},
		{
			// 场景 E 的真实 JSON（来自 leetcode.cn API 返回）
			name:     "#278 first-bad-version (正常 manual)",
			jsonStr:  `{"name":"firstBadVersion","params":[{"name":"n","type":"integer"}],"return":{"type":"integer"},"manual":true}`,
			expected: "integer",
		},
		{
			// 场景 A 的真实 JSON
			name:     "#1 twoSum (正常题目)",
			jsonStr:  `{"name":"twoSum","params":[{"name":"nums","type":"integer[]"},{"name":"target","type":"integer"}],"return":{"type":"integer[]","size":2}}`,
			expected: "integer[]",
		},
		{
			// 场景 B 的真实 JSON — return=void + output
			// 找一个真实的 SD 题目: #855 ExamRoom
			// 简化版只关注关键字段
			name:     "ExamRoom (SD, return=void+output)",
			jsonStr:  `{"name":"ExamRoom","params":[{"name":"inputs","type":"integer[]"},{"name":"inputs","type":"integer[]"}],"return":{"type":"void"},"output":{"paramindex":1},"systemdesign":true,"classname":"ExamRoom","constructor":{"params":[{"name":"n","type":"integer"}]},"methods":[{"name":"seat","params":[],"return":{"type":"integer"}},{"name":"leave","params":[{"name":"p","type":"integer"}],"return":{"type":"void"}}]}`,
			expected: "integer[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 UnmarshalJSON 的实际调用路径：
			// 1. json.Unmarshal 到 MetaData 结构体
			var m MetaData
			// 注意：真实流程中 metaData 是字符串类型，会先 Unmarshal 到 string，
			// 再二次 Unmarshal。这里直接用结构体 Unmarshal 等价。
			if err := unmarshalMetaData([]byte(tt.jsonStr), &m); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			// 2. normalize() 由 UnmarshalJSON 内部调用

			got := m.ResultType()
			if got != tt.expected {
				t.Errorf("ResultType() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// unmarshalMetaData 模拟 UnmarshalJSON 的完整流程（不走字符串二次解引用，直接 JSON → struct）
func unmarshalMetaData(data []byte, m *MetaData) error {
	type alias MetaData
	if err := json.Unmarshal(data, (*alias)(m)); err != nil {
		return err
	}
	m.normalize()
	return nil
}
