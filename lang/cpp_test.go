package lang

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/j178/leetgo/leetcode"
	cppUtils "github.com/j178/leetgo/testutils/cpp"
)

func TestCppInitWorkspaceRefreshesHeader(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("LEETGO_HOME", homeDir)
	if err := os.Mkdir(filepath.Join(homeDir, "cache"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := UpdateDep(cppGen); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()
	headerPath := filepath.Join(outDir, cppUtils.HeaderName)
	if err := os.WriteFile(headerPath, []byte("old header"), 0o600); err != nil {
		t.Fatal(err)
	}

	bitsDir := filepath.Join(outDir, "bits")
	if err := os.Mkdir(bitsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	stdCxxPath := filepath.Join(bitsDir, "stdc++.h")
	if err := os.WriteFile(stdCxxPath, cppUtils.StdCxxContent, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := cppGen.InitWorkspace(outDir); err != nil {
		t.Fatal(err)
	}

	header, err := os.ReadFile(headerPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(header, cppUtils.HeaderContent) {
		t.Fatal("LC_IO.h was not updated")
	}
}

// buildVoidQuestion returns a minimal QuestionData for a void in-place problem
// that mirrors problem 283 (moveZeroes): void moveZeroes(vector<int>& nums)
func buildVoidQuestion() *leetcode.QuestionData {
	return &leetcode.QuestionData{
		TitleSlug: "move-zeroes",
		MetaData: leetcode.MetaData{
			Name: "moveZeroes",
			Params: []leetcode.MetaDataParam{
				{Name: "nums", Type: "integer[]"},
			},
			Return: &leetcode.MetaDataReturn{Type: "void"},
			Output: &leetcode.MetaDataOutput{ParamIndex: 0},
		},
	}
}

// buildNonVoidQuestion returns a minimal QuestionData for a problem that returns a value
// that mirrors problem 1 (twoSum): vector<int> twoSum(vector<int>& nums, int target)
func buildNonVoidQuestion() *leetcode.QuestionData {
	return &leetcode.QuestionData{
		TitleSlug: "two-sum",
		MetaData: leetcode.MetaData{
			Name: "twoSum",
			Params: []leetcode.MetaDataParam{
				{Name: "nums", Type: "integer[]"},
				{Name: "target", Type: "integer"},
			},
			Return: &leetcode.MetaDataReturn{Type: "integer[]"},
		},
	}
}

func assertSnapshot(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("snapshot mismatch:\n--- want\n%s\n--- got\n%s", want, got)
	}
}

func TestGenerateCallCode_VoidReturn(t *testing.T) {
	assertSnapshot(t, (cpp{}).generateCallCode(buildVoidQuestion()), `	obj.moveZeroes(nums);
`)
}

func TestGenerateCallCode_NonVoidReturn(t *testing.T) {
	assertSnapshot(t, (cpp{}).generateCallCode(buildNonVoidQuestion()), `	auto res = obj.twoSum(nums, target);
`)
}

func TestGeneratePrintCode_VoidReturn_WithOutput(t *testing.T) {
	assertSnapshot(t, (cpp{}).generatePrintCode(buildVoidQuestion()), `	LeetCodeIO::print(out_stream, nums);
	cout << "\noutput: " << out_stream.rdbuf() << '\n';
`)
}

// buildVoidNoOutputQuestion returns a minimal QuestionData for a void method
// with no Output metadata, exercising the "null" fallback branch.
func buildVoidNoOutputQuestion() *leetcode.QuestionData {
	return &leetcode.QuestionData{
		TitleSlug: "void-no-output",
		MetaData: leetcode.MetaData{
			Name: "doSomething",
			Params: []leetcode.MetaDataParam{
				{Name: "x", Type: "integer"},
			},
			Return: &leetcode.MetaDataReturn{Type: "void"},
			// Output is nil — should trigger "null" fallback
		},
	}
}

func TestGeneratePrintCode_VoidReturn_NoOutput(t *testing.T) {
	assertSnapshot(t, (cpp{}).generatePrintCode(buildVoidNoOutputQuestion()), `	out_stream << "null";
	cout << "\noutput: " << out_stream.rdbuf() << '\n';
`)
}

func TestGeneratePrintCode_NonVoidReturn(t *testing.T) {
	assertSnapshot(t, (cpp{}).generatePrintCode(buildNonVoidQuestion()), `	LeetCodeIO::print(out_stream, res);
	cout << "\noutput: " << out_stream.rdbuf() << '\n';
`)
}

func buildSystemDesignQuestion() *leetcode.QuestionData {
	return &leetcode.QuestionData{
		TitleSlug: "counter",
		MetaData: leetcode.MetaData{
			SystemDesign: true,
			ClassName:    "Counter",
			Constructor: leetcode.MetaDataConstructor{
				Params: []leetcode.MetaDataParam{{Name: "start", Type: "integer"}},
			},
			Methods: []leetcode.MetaDataMethod{
				{
					Name: "add",
					Params: []leetcode.MetaDataParam{
						{Name: "left", Type: "integer"},
						{Name: "right", Type: "integer"},
					},
					Return: leetcode.MetaDataReturn{Type: "void"},
				},
				{
					Name:   "addAll",
					Params: []leetcode.MetaDataParam{{Name: "values", Type: "integer[]"}},
					Return: leetcode.MetaDataReturn{Type: "void"},
				},
				{
					Name:   "value",
					Return: leetcode.MetaDataReturn{Type: "integer"},
				},
			},
		},
	}
}

const nonSystemDesignMainSnapshot = `int main() {
	ios_base::sync_with_stdio(false);
	try {
		vector<int> nums = LeetCodeIO::deserialize<vector<int>>(cin);
		int target = LeetCodeIO::deserialize<int>(cin);

		Solution obj;
		auto res = obj.twoSum(nums, target);

		stringstream out_stream;
		LeetCodeIO::print(out_stream, res);
		cout << "\noutput: " << out_stream.rdbuf() << '\n';
	} catch (const LeetCodeIO::Error &error) {
		cerr << "LC_IO: " << error.what() << '\n';
		return 2;
	}
	return 0;
}`

const systemDesignMainSnapshot = `int main() {
	ios_base::sync_with_stdio(false);
	try {
		auto method_names = LeetCodeIO::deserialize<vector<string>>(cin);
		auto params = LeetCodeIO::split_array(cin);

		unique_ptr<Counter> obj;
		vector<string> output;
		const unordered_map<string, function<void(const vector<string> &)>> methods = {
			{ "Counter", [&](const vector<string> &method_params) {
				int start = LeetCodeIO::deserialize<int>(method_params, 0);
				obj = make_unique<Counter>(start);
				output.push_back("null");
			} },
			{ "add", [&](const vector<string> &method_params) {
				int left = LeetCodeIO::deserialize<int>(method_params, 0);
				int right = LeetCodeIO::deserialize<int>(method_params, 1);
				obj->add(left, right);
				output.push_back("null");
			} },
			{ "addAll", [&](const vector<string> &method_params) {
				vector<int> values = LeetCodeIO::deserialize<vector<int>>(method_params, 0);
				obj->addAll(values);
				output.push_back("null");
			} },
			{ "value", [&](const vector<string> &) {
				output.push_back(LeetCodeIO::serialize(obj->value()));
			} },
		};
		if (method_names.size() != params.size()) {
			throw LeetCodeIO::Error("method and parameter counts differ");
		}
		for (size_t i = 0; i < method_names.size(); ++i) {
			auto method_params = LeetCodeIO::split_array(params[i]);
			methods.at(method_names[i])(method_params);
		}
		cout << "\noutput: " << LeetCodeIO::join_array(output) << '\n';
	} catch (const LeetCodeIO::Error &error) {
		cerr << "LC_IO: " << error.what() << '\n';
		return 2;
	}
	return 0;
}`

func compileCppProgram(t *testing.T, source string) string {
	t.Helper()

	compiler, err := exec.LookPath("g++")
	if err != nil {
		t.Skip("g++ is not installed")
	}

	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "main.cpp")
	binaryPath := filepath.Join(dir, "main")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, cppUtils.HeaderName), cppUtils.HeaderContent, 0o600); err != nil {
		t.Fatal(err)
	}
	bitsDir := filepath.Join(dir, "bits")
	if err := os.Mkdir(bitsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bitsDir, "stdc++.h"), cppUtils.StdCxxContent, 0o600); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(compiler, "-std=c++17", "-O2", "-Wall", "-Wextra", "-Wpedantic", "-Werror", "-I", dir, sourcePath, "-o", binaryPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("failed to compile generated C++:\n%s", output)
	}
	return binaryPath
}

func runCppProgram(t *testing.T, binary, input string) (string, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, binary)
	command.Stdin = strings.NewReader(input)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("generated C++ program timed out: %v", ctx.Err())
	}
	return string(output), err
}

func TestGeneratedCppProgram_NonSystemDesign(t *testing.T) {
	mainCode, err := (cpp{}).generateTestContent(buildNonVoidQuestion())
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, mainCode, nonSystemDesignMainSnapshot)
	source := `#include <bits/stdc++.h>
#include "LC_IO.h"
using namespace std;

class Solution {
public:
	vector<int> twoSum(vector<int> &nums, int target) {
		for (size_t i = 0; i < nums.size(); ++i) {
			for (size_t j = i + 1; j < nums.size(); ++j) {
				if (nums[i] + nums[j] == target) return {static_cast<int>(i), static_cast<int>(j)};
			}
		}
		return {};
	}
};

` + mainCode
	binary := compileCppProgram(t, source)

	output, err := runCppProgram(t, binary, "[2,7,11,15]\n9\n")
	if err != nil {
		t.Fatalf("generated C++ failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "output: [0,1]") {
		t.Fatalf("unexpected generated C++ output: %q", output)
	}

	output, err = runCppProgram(t, binary, "[2,7\n9\n")
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 2 {
		t.Fatalf("malformed input should exit with code 2, got %v\n%s", err, output)
	}
	if !strings.Contains(output, "LC_IO:") {
		t.Fatalf("malformed input should include a diagnostic, got %q", output)
	}
}

func TestGeneratedCppProgram_SystemDesign(t *testing.T) {
	mainCode, err := (cpp{}).generateTestContent(buildSystemDesignQuestion())
	if err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, mainCode, systemDesignMainSnapshot)
	source := `#include <bits/stdc++.h>
#include "LC_IO.h"
using namespace std;

class Counter {
public:
	explicit Counter(int start) : current(start) {}
	void add(int left, int right) { current += left + right; }
	void addAll(const vector<int> &values) {
		for (int value : values) current += value;
	}
	int value() const { return current; }

private:
	int current;
};

` + mainCode
	binary := compileCppProgram(t, source)

	input := "[\"Counter\",\"add\",\"addAll\",\"value\"]\n[[5],[2,3],[[4,5]],[]]\n"
	output, err := runCppProgram(t, binary, input)
	if err != nil {
		t.Fatalf("generated system design C++ failed: %v\n%s", err, output)
	}
	if !strings.Contains(output, "output: [null,null,null,19]") {
		t.Fatalf("unexpected generated system design output: %q", output)
	}
}

func TestCppHeaderSupportsLegacyGeneratedCode(t *testing.T) {
	// This mirrors the system design harness generated before the LC_IO rewrite.
	source := `#include <bits/stdc++.h>
#include "LC_IO.h"
using namespace std;

class Counter {
public:
	explicit Counter(int start) : current(start) {}
	void add(int left, int right) { current += left + right; }
	void addAll(const vector<int> &values) {
		for (int value : values) current += value;
	}
	int value() const { return current; }

private:
	int current;
};

int main() {
	ios_base::sync_with_stdio(false);
	stringstream out_stream;

	vector<string> method_names;
	LeetCodeIO::scan(cin, method_names);

	Counter *obj;
	const unordered_map<string, function<void()>> methods = {
		{ "Counter", [&]() {
			int start; LeetCodeIO::scan(cin, start); cin.ignore();
			obj = new Counter(start);
			out_stream << "null,";
		} },
		{ "add", [&]() {
			int left; LeetCodeIO::scan(cin, left); cin.ignore();
			int right; LeetCodeIO::scan(cin, right); cin.ignore();
			obj->add(left, right);
			out_stream << "null,";
		} },
		{ "addAll", [&]() {
			vector<int> values; LeetCodeIO::scan(cin, values); cin.ignore();
			obj->addAll(values);
			out_stream << "null,";
		} },
		{ "value", [&]() {
			cin.ignore();
			LeetCodeIO::print(out_stream, obj->value()); out_stream << ',';
		} },
	};
	cin >> ws;
	out_stream << '[';
	for (auto &&method_name : method_names) {
		cin.ignore(2);
		methods.at(method_name)();
	}
	cin.ignore();
	out_stream.seekp(-1, ios_base::end); out_stream << ']';
	cout << "\noutput: " << out_stream.rdbuf() << endl;
	delete obj;
	return 0;
}`
	binary := compileCppProgram(t, source)

	input := "[\"Counter\",\"add\",\"addAll\",\"value\"]\n[[5],[2,3],[[4,5]],[]]\n"
	output, err := runCppProgram(t, binary, input)
	if err != nil {
		t.Fatalf("legacy generated C++ failed with the new LC_IO.h: %v\n%s", err, output)
	}
	if !strings.Contains(output, "output: [null,null,null,19]") {
		t.Fatalf("unexpected legacy generated C++ output: %q", output)
	}
}
