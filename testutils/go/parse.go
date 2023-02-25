package goutils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func MustSplitArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	splits, err := SplitArray(raw)
	if err != nil {
		panic(err)
	}
	return splits
}

func SplitArray(raw string) (splits []string, err error) {
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

// DeserializeValue deserialize a string to a reflect.Value
func DeserializeValue(ty reflect.Type, raw string) (reflect.Value, error) {
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
		splits, err := SplitArray(raw)
		if err != nil {
			return z, fmt.Errorf("invalid array: %s", raw)
		}
		sl := reflect.MakeSlice(ty, 0, len(splits))
		for _, s := range splits {
			e, err := DeserializeValue(ty.Elem(), s)
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
	return z, fmt.Errorf("unknown type %s", ty.String())
}

// Deserialize deserialize a string to a type.
func Deserialize[T any](raw string) T {
	raw = strings.TrimSpace(raw)
	var z T
	ty := reflect.TypeOf(z)
	v, err := DeserializeValue(ty, raw)
	if err != nil {
		panic(fmt.Errorf("deserialize failed: %w", err))
	}
	rv := reflect.ValueOf(&z)
	rv.Elem().Set(v)
	return z
}

func serialize(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Slice:
		sb := &strings.Builder{}
		sb.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				sb.WriteByte(',')
			}
			_s, err := serialize(v.Index(i))
			if err != nil {
				return "", err
			}
			sb.WriteString(_s)
		}
		sb.WriteByte(']')
		return sb.String(), nil
	case reflect.Ptr: // *TreeNode, *ListNode
		switch tpName := v.Type().Elem().Name(); tpName {
		case "TreeNode":
			return v.Interface().(*TreeNode).ToString(), nil
		case "ListNode":
			return v.Interface().(*ListNode).ToString(), nil
		default:
			return "", fmt.Errorf("unknown type %s", tpName)
		}
	case reflect.String:
		return fmt.Sprintf(`"%s"`, v.Interface()), nil
	case reflect.Uint8: // byte
		return fmt.Sprintf(`"%c"`, v.Interface()), nil
	case reflect.Float64:
		return fmt.Sprintf(`%.5f`, v.Interface()), nil
	default: // int uint int64 uint64 bool
		return fmt.Sprintf(`%v`, v.Interface()), nil
	}
}

func Serialize(v any) string {
	vt := reflect.ValueOf(v)
	s, err := serialize(vt)
	if err != nil {
		panic(fmt.Errorf("serialize failed: %w", err))
	}
	return s
}
