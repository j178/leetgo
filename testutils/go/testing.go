package common

/*
Much appreciated to EndlessCheng
Copy from https://github.com/EndlessCheng/codeforces-go/blob/ae5b312f3f/leetcode/testutil/leetcode.go
*/

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func parseRawArray(rawArray string) (splits []string, err error) {
	invalidErr := fmt.Errorf("invalid test data: %s", rawArray)

	// check [] at leftmost and rightmost
	if len(rawArray) <= 1 || rawArray[0] != '[' || rawArray[len(rawArray)-1] != ']' {
		return nil, invalidErr
	}

	// ignore [] at leftmost and rightmost
	rawArray = rawArray[1 : len(rawArray)-1]
	if rawArray == "" {
		return
	}

	var depth, quote int
	for i := 0; i < len(rawArray); {
		j := i
	outer:
		for ; j < len(rawArray); j++ {
			switch rawArray[j] {
			case '[':
				depth++
			case ']':
				depth--
			case '"':
				quote++
			case ',':
				if depth == 0 && quote%2 == 0 {
					break outer
				}
			}
		}
		splits = append(splits, strings.TrimSpace(rawArray[i:j]))
		i = j + 1 // skip sep
	}
	if depth != 0 || quote%2 != 0 {
		return nil, invalidErr
	}
	return
}

func parseRawArg(tp reflect.Type, rawData string) (v reflect.Value, err error) {
	rawData = strings.TrimSpace(rawData)
	invalidErr := fmt.Errorf("invalid test data: %s", rawData)
	switch tp.Kind() {
	case reflect.String:
		if len(rawData) <= 1 || rawData[0] != '"' && rawData[0] != '\'' || rawData[len(rawData)-1] != rawData[0] {
			return reflect.Value{}, invalidErr
		}
		// 处理转义字符
		w := strings.Builder{}
		// remove " (or ') at leftmost and rightmost
		for i := 1; i < len(rawData)-1; i++ {
			switch rawData[i] {
			case '\\':
				i++
				switch rawData[i] {
				case '"':
					w.WriteByte('"')
				case '\\':
					w.WriteByte('\\')
				case '/':
					w.WriteByte('/')
				case 'b':
					w.WriteByte('\b')
				case 'f':
					w.WriteByte('\f')
				case 'n':
					w.WriteByte('\n')
				case 't':
					w.WriteByte('\t')
				default:
					w.WriteByte('\\')
					w.WriteByte(rawData[i])
				}
			default:
				w.WriteByte(rawData[i])
			}
		}
		v = reflect.ValueOf(w.String())
	case reflect.Uint8: // byte
		// rawData like "a" or 'a'
		if len(rawData) != 3 || rawData[0] != '"' && rawData[0] != '\'' || rawData[2] != rawData[0] {
			return reflect.Value{}, invalidErr
		}
		v = reflect.ValueOf(rawData[1])
	case reflect.Int:
		i, er := strconv.Atoi(rawData)
		if er != nil {
			return reflect.Value{}, invalidErr
		}
		v = reflect.ValueOf(i)
	case reflect.Uint:
		i, er := strconv.Atoi(rawData)
		if er != nil {
			return reflect.Value{}, invalidErr
		}
		v = reflect.ValueOf(uint(i))
	case reflect.Int64:
		i, er := strconv.Atoi(rawData)
		if er != nil {
			return reflect.Value{}, invalidErr
		}
		v = reflect.ValueOf(int64(i))
	case reflect.Uint64:
		i, er := strconv.Atoi(rawData)
		if er != nil {
			return reflect.Value{}, invalidErr
		}
		v = reflect.ValueOf(uint64(i))
	case reflect.Float64:
		f, er := strconv.ParseFloat(rawData, 64)
		if er != nil {
			return reflect.Value{}, invalidErr
		}
		v = reflect.ValueOf(f)
	case reflect.Bool:
		b, er := strconv.ParseBool(rawData)
		if er != nil {
			return reflect.Value{}, invalidErr
		}
		v = reflect.ValueOf(b)
	case reflect.Slice:
		splits, er := parseRawArray(rawData)
		if er != nil {
			return reflect.Value{}, er
		}
		v = reflect.New(tp).Elem()
		for _, s := range splits {
			_v, er := parseRawArg(tp.Elem(), s)
			if er != nil {
				return reflect.Value{}, er
			}
			v = reflect.Append(v, _v)
		}
	case reflect.Ptr: // *TreeNode, *ListNode, *Point, *Interval
		switch tpName := tp.Elem().Name(); tpName {
		case "TreeNode":
			root, er := DeserializeTreeNode(rawData)
			if er != nil {
				return reflect.Value{}, er
			}
			v = reflect.ValueOf(root)
		case "ListNode":
			head, er := DeserializeListNode(rawData)
			if er != nil {
				return reflect.Value{}, er
			}
			v = reflect.ValueOf(head)
		default:
			return reflect.Value{}, fmt.Errorf("unknown type %s", tpName)
		}
	default:
		return reflect.Value{}, fmt.Errorf("unknown type %s", tp.Name())
	}
	return
}

func toRawString(v reflect.Value) (s string, err error) {
	switch v.Kind() {
	case reflect.Slice:
		sb := &strings.Builder{}
		sb.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				sb.WriteByte(',')
			}
			_s, er := toRawString(v.Index(i))
			if er != nil {
				return "", er
			}
			sb.WriteString(_s)
		}
		sb.WriteByte(']')
		s = sb.String()
	case reflect.Ptr: // *TreeNode, *ListNode, *Point, *Interval
		switch tpName := v.Type().Elem().Name(); tpName {
		case "TreeNode":
			s = v.Interface().(*TreeNode).ToString()
		case "ListNode":
			s = v.Interface().(*ListNode).ToString()
		default:
			return "", fmt.Errorf("unknown type %s", tpName)
		}
	case reflect.String:
		s = fmt.Sprintf(`"%s"`, v.Interface())
	case reflect.Uint8: // byte
		s = fmt.Sprintf(`"%c"`, v.Interface())
	case reflect.Float64:
		s = fmt.Sprintf(`%.5f`, v.Interface())
	default: // int uint int64 uint64 bool
		s = fmt.Sprintf(`%v`, v.Interface())
	}
	return
}

// rawExamples[i] = 输入+输出
// 若反射出来的函数或 rawExamples 数据不合法，则会返回一个非空的 error，否则返回 nil

func RunTests(t *testing.T, f interface{}, examples [][]string, targetCaseNum int) (err error) {
	t.Helper()

	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)
	if fType.Kind() != reflect.Func {
		return errors.New("f must be a function")
	}
	for caseNo, example := range examples {
		if len(example) != fType.NumIn()+fType.NumOut() {
			return fmt.Errorf(
				"Case#%d invalid: len(example) = %d, but we need %d+%d",
				caseNo+1,
				len(example),
				fType.NumIn(),
				fType.NumOut(),
			)
		}
	}

	test := func(caseNo int, example []string) (passed bool, err error) {
		rawIn := example[:fType.NumIn()]
		ins := make([]reflect.Value, len(rawIn))
		for i, rawArg := range rawIn {
			rawArg = trimSpaceAndNewLine(rawArg)
			ins[i], err = parseRawArg(fType.In(i), rawArg)
			if err != nil {
				return
			}
		}
		// just check rawExpectedOuts is valid or not
		rawExpectedOuts := example[fType.NumIn():]
		for i := range rawExpectedOuts {
			rawExpectedOuts[i] = trimSpaceAndNewLine(rawExpectedOuts[i])
			if _, err = parseRawArg(fType.Out(i), rawExpectedOuts[i]); err != nil {
				return
			}
		}

		const maxInputSize = 150
		inputInfo := strings.Join(rawIn, "\n")
		if len(inputInfo) > maxInputSize { // 截断过长的输入
			inputInfo = inputInfo[:maxInputSize] + "..."
		}

		subTestName := fmt.Sprintf("Case#%d", caseNo)
		passed = true
		t.Run(
			subTestName, func(t *testing.T) {
				var outs []reflect.Value
				if isTLE(func() { outs = fValue.Call(ins) }) {
					t.Errorf(
						"Time Limit Exceeded\n"+
							"input: %s", inputInfo,
					)
					return
				}

				for i, out := range outs {
					rawActualOut, er := toRawString(out)
					if er != nil {
						t.Fatalf("Convert result to string error: %v", er)
					}
					if AssertOutput && rawActualOut != rawExpectedOuts[i] {
						t.Errorf(
							"Not equal\n"+
								"expected: %s\n"+
								"actual  : %s\n"+
								"input   : %s",
							rawExpectedOuts[i],
							rawActualOut,
							inputInfo,
						)
						passed = false
					}
				}
			},
		)
		return
	}

	// 例如，-1 表示最后一个测试用例
	if targetCaseNum < 0 {
		targetCaseNum += len(examples) + 1
	}
	if targetCaseNum > 0 {
		example := examples[targetCaseNum-1]
		passed, err := test(targetCaseNum, example)
		if err != nil {
			return err
		}
		if !passed {
			return nil
		}
		t.Logf("Case#%d passed, continue to run all tests", targetCaseNum)
	}

	// 如果测试的是单个用例，而且通过了，则继续跑一遍全量用例
	for curCaseNum, example := range examples {
		// 跳过已经跑过的目标测试
		if targetCaseNum > 0 && curCaseNum+1 == targetCaseNum {
			continue
		}
		_, err = test(curCaseNum+1, example)
		if err != nil {
			return
		}
	}
	return nil
}

func RunTestsWithString(t *testing.T, f interface{}, testcases string, targetCaseNum int) (err error) {
	lines := parseTestCases(testcases)
	n := len(lines)
	if n == 0 {
		return errors.New("invalid testcases: empty testcases")
	}

	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		return errors.New("f must be a function")
	}
	// 每 fNumIn+fNumOut 行一组数据
	fNumIn := fType.NumIn()
	fNumOut := fType.NumOut()
	tcSize := fNumIn + fNumOut
	if n%tcSize != 0 {
		return fmt.Errorf("invalid testcases: got %d lines, should be a multiple of %d", n, tcSize)
	}

	examples := make([][]string, 0, n/tcSize)
	for i := 0; i < n; i += tcSize {
		examples = append(examples, lines[i:i+tcSize])
	}
	return RunTests(t, f, examples, targetCaseNum)
}

func RunClassTests(t *testing.T, constructor interface{}, examples [][3]string, targetCaseNum int) (err error) {
	cType := reflect.TypeOf(constructor)
	cFunc := reflect.ValueOf(constructor)
	if cType.Kind() != reflect.Func {
		return fmt.Errorf("constructor must be a function")
	}
	if cType.NumOut() != 1 {
		return fmt.Errorf("constructor must have one and only one return value")
	}

	test := func(caseNo int, example [3]string) (passed bool, err error) {
		names := strings.TrimSpace(example[0])
		inputArgs := strings.TrimSpace(example[1])
		rawExpectedOut := strings.TrimSpace(example[2])
		// 去除 rawExpectedOut 中逗号后的空格
		rawExpectedOut = strings.ReplaceAll(rawExpectedOut, ", ", ",")

		methodNames, err := parseRawArray(names)
		if err != nil {
			return
		}
		for i, name := range methodNames {
			name = name[1 : len(name)-1] // 移除引号
			name = strings.Title(name)   // 首字母大写以匹配模板方法名称
			methodNames[i] = name
		}

		rawArgsList, err := parseRawArray(inputArgs)
		if err != nil {
			return
		}
		if len(rawArgsList) != len(methodNames) {
			return false, fmt.Errorf(
				"Case#%d invalid: mismatch names and input args (%d != %d)",
				caseNo,
				len(methodNames),
				len(rawArgsList),
			)
		}

		constructorArgs, err := parseRawArray(rawArgsList[0])
		if err != nil {
			return
		}
		constructorIns := make([]reflect.Value, len(constructorArgs))
		for i, arg := range constructorArgs {
			constructorIns[i], err = parseRawArg(cType.In(i), arg)
			if err != nil {
				return
			}
		}

		obj := cFunc.Call(constructorIns)[0]
		// use a pointer to call methods
		pObj := reflect.New(obj.Type())
		pObj.Elem().Set(obj)

		if DebugCallIndex < 0 {
			DebugCallIndex += len(rawArgsList)
		}
		passed = true
		subTestName := fmt.Sprintf("Case#%d", caseNo)
		t.Run(
			subTestName, func(t *testing.T) {
				rawActualOut := strings.Builder{}
				rawActualOut.WriteString("[null")

				for callIndex := 1; callIndex < len(rawArgsList); callIndex++ {
					name := methodNames[callIndex]
					method := pObj.MethodByName(name)
					emptyValue := reflect.Value{}
					if method == emptyValue {
						t.Fatalf("invalid test data: method %s not exist", methodNames[callIndex])
					}
					methodType := method.Type()

					// parse method input
					methodArgs, err := parseRawArray(rawArgsList[callIndex])
					if err != nil {
						t.Fatalf("invalid test data: invalid input: %v", err)
					}
					in := make([]reflect.Value, methodType.NumIn()) // 注意：若入参为空，methodArgs 可能是 [] 也可能是 [null]
					for i := range in {
						in[i], err = parseRawArg(methodType.In(i), methodArgs[i])
						if err != nil {
							t.Fatalf("invalid test data: invalid input: %v", err)
						}
					}

					if callIndex == DebugCallIndex {
						print() // 在这里打断点
					}

					// call method
					var actualOuts []reflect.Value
					_f := func() { actualOuts = method.Call(in) }
					if isTLE(_f) {
						t.Errorf(
							"Time Limit Exceeded\n"+
								"Call Index: %d\n"+
								"Func      : %s(%s)",
							callIndex,
							name,
							rawArgsList[callIndex][1:len(rawArgsList[callIndex])-1],
						)
						return
					}

					if len(actualOuts) > 0 {
						s, err := toRawString(actualOuts[0])
						if err != nil {
							return
						}
						rawActualOut.WriteByte(',')
						rawActualOut.WriteString(s)
					} else {
						rawActualOut.WriteString(",null")
					}
				}
				rawActualOut.WriteByte(']')
				// todo: 提示错在哪个 callIndex 上
				if AssertOutput && rawExpectedOut != rawActualOut.String() {
					t.Errorf(
						"Not equal\n"+
							"expected: %s\n"+
							"actual  : %s\n",
						rawExpectedOut,
						rawActualOut.String(),
					)
					passed = false
				}
			},
		)
		return
	}

	// 例如，-1 表示最后一个测试用例
	if targetCaseNum < 0 {
		targetCaseNum += len(examples) + 1
	}
	if targetCaseNum > 0 {
		example := examples[targetCaseNum-1]
		passed, err := test(targetCaseNum, example)
		if err != nil {
			return err
		}
		if !passed {
			return nil
		}
		t.Logf("Case#%d passed, continue to run all tests", targetCaseNum)
	}

	// 如果测试的是单个用例，而且通过了，则继续跑一遍全量用例
	for curCaseNum, example := range examples {
		// 跳过已经跑过的目标测试
		if targetCaseNum > 0 && curCaseNum+1 == targetCaseNum {
			continue
		}
		_, err = test(curCaseNum+1, example)
		if err != nil {
			return
		}
	}

	return nil
}

func RunClassTestsWithString(t *testing.T, constructor interface{}, testcases string, targetCaseNum int) error {
	lines := parseTestCases(testcases)
	n := len(lines)
	if n == 0 {
		return fmt.Errorf("invalid testcases: empty testcases")
	}

	// 每三行一组数据
	if n%3 != 0 {
		return fmt.Errorf("invalid testcases: got %d lines, should be a multiple of 3", n)
	}

	examples := make([][3]string, 0, n/3)
	for i := 0; i < n; i += 3 {
		examples = append(examples, [3]string{lines[i], lines[i+1], lines[i+2]})
	}
	return RunClassTests(t, constructor, examples, targetCaseNum)
}
