package goutils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func splitArray(raw string) (splits []string, err error) {
	invalidErr := fmt.Errorf("invalid array: %s", raw)

	// check [] at leftmost and rightmost
	if len(raw) <= 1 || raw[0] != '[' || raw[len(raw)-1] != ']' {
		return nil, invalidErr
	}

	// ignore [] at leftmost and rightmost
	raw = raw[1 : len(raw)-1]
	if raw == "" {
		return
	}

	var depth, quote int
	for i := 0; i < len(raw); {
		j := i
	outer:
		for ; j < len(raw); j++ {
			switch raw[j] {
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
		splits = append(splits, strings.TrimSpace(raw[i:j]))
		i = j + 1 // skip sep
	}
	if depth != 0 || quote%2 != 0 {
		return nil, invalidErr
	}
	return
}

type GoTypeName string

func typeNameToType(ty GoTypeName) reflect.Type {
	switch ty {
	case "int":
		return reflect.TypeOf(0)
	case "float64":
		return reflect.TypeOf(float64(0))
	case "string":
		return reflect.TypeOf("")
	case "bool":
		return reflect.TypeOf(false)
	case "byte":
		return reflect.TypeOf(byte(0))
	case "*TreeNode":
		return reflect.TypeOf((*TreeNode)(nil))
	case "*ListNode":
		return reflect.TypeOf((*ListNode)(nil))
	default:
		if strings.HasPrefix(string(ty), "[]") {
			et := typeNameToType(GoTypeName(string(ty)[2:]))
			if et == nil {
				return nil
			}
			return reflect.SliceOf(et)
		}
	}
	return nil
}

func DeserializeByGoType(tpName GoTypeName, raw string) (reflect.Value, error) {
	raw = strings.TrimSpace(raw)
	ty := typeNameToType(tpName)
	if ty == nil {
		return reflect.Value{}, fmt.Errorf("invalid type: %s", tpName)
	}
	return deserialize(ty, raw)
}

func deserialize(ty reflect.Type, raw string) (reflect.Value, error) {
	z := reflect.Value{}
	switch ty.Kind() {
	case reflect.Bool:
		if raw != "true" && raw != "false" {
			return z, fmt.Errorf("invalid bool: %s", raw)
		}
		b := raw == "true"
		return reflect.ValueOf(b), nil
	case reflect.Uint8: // byte
		if len(raw) != 3 || raw[0] != '"' && raw[0] != '\'' || raw[2] != raw[0] {
			return z, fmt.Errorf("invalid byte: %s", raw)
		}
		return reflect.ValueOf(raw[1]), nil
	case reflect.String:
		s, err := strconv.Unquote(raw)
		if err != nil {
			return z, fmt.Errorf("invalid string: %s", raw)
		}
		return reflect.ValueOf(s), nil
	case reflect.Int, reflect.Int32:
		i, err := strconv.Atoi(raw)
		if err != nil {
			return z, fmt.Errorf("invalid int: %s", raw)
		}
		return reflect.ValueOf(i), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return z, fmt.Errorf("invalid int64: %s", raw)
		}
		return reflect.ValueOf(i), nil
	case reflect.Uint, reflect.Uint32:
		i, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return z, fmt.Errorf("invalid uint: %s", raw)
		}
		return reflect.ValueOf(uint(i)), nil
	case reflect.Uint64:
		i, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return z, fmt.Errorf("invalid uint64: %s", raw)
		}
		return reflect.ValueOf(i), nil
	case reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return z, fmt.Errorf("invalid float64: %s", raw)
		}
		return reflect.ValueOf(f), nil
	case reflect.Slice:
		splits, err := splitArray(raw)
		if err != nil {
			return z, fmt.Errorf("invalid array: %s", raw)
		}
		sl := reflect.MakeSlice(ty, 0, len(splits))
		for _, s := range splits {
			e, err := deserialize(ty.Elem(), s)
			if err != nil {
				return z, err
			}
			sl = reflect.Append(sl, e)
		}
		return sl, nil
	case reflect.Ptr:
		switch ty.Elem().Name() {
		case "TreeNode":
			root, err := DeserializeTreeNode(raw)
			if err != nil {
				return z, err
			}
			return reflect.ValueOf(root), nil
		case "ListNode":
			head, err := DeserializeListNode(raw)
			if err != nil {
				return z, err
			}
			return reflect.ValueOf(head), nil
		}
	}
	return z, fmt.Errorf("unknown type %s", ty.Name())
}

func Deserialize[T any](raw string) T {
	raw = strings.TrimSpace(raw)
	var z T
	ty := reflect.TypeOf(z)
	v, err := deserialize(ty, raw)
	if err != nil {
		panic(fmt.Errorf("deserialize %s failed: %w", raw, err))
	}
	rv := reflect.ValueOf(&z)
	rv.Elem().Set(v)
	return z
}

func serialize(v reflect.Value) (s string, err error) {
	switch v.Kind() {
	case reflect.Slice:
		sb := &strings.Builder{}
		sb.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				sb.WriteByte(',')
			}
			_s, er := serialize(v.Index(i))
			if er != nil {
				return "", er
			}
			sb.WriteString(_s)
		}
		sb.WriteByte(']')
		s = sb.String()
	case reflect.Ptr: // *TreeNode, *ListNode
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

func Serialize(v any) string {
	vt := reflect.ValueOf(v)
	s, err := serialize(vt)
	if err != nil {
		panic(fmt.Errorf("serialize %v failed: %w", v, err))
	}
	return s
}
