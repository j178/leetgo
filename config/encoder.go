package config

import (
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Code below is copied from https://github.com/siderolabs/talos/blob/main/pkg/machinery/config/encoder/encoder.go
// with some modifications.

func isEmpty(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Ptr:
		return value.IsNil()
	case reflect.Map:
		return len(value.MapKeys()) == 0
	case reflect.Slice:
		return value.Len() == 0
	default:
		return value.IsZero()
	}
}

func isNil(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	//nolint:exhaustive
	switch value.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func toYamlNode(in interface{}) (*yaml.Node, error) {
	node := &yaml.Node{}

	// do not wrap yaml.Node into yaml.Node
	if n, ok := in.(*yaml.Node); ok {
		return n, nil
	}

	// if input implements yaml.Marshaler we should use that marshaller instead
	// same way as regular yaml marshal does
	if m, ok := in.(yaml.Marshaler); ok && !isNil(reflect.ValueOf(in)) {
		res, err := m.MarshalYAML()
		if err != nil {
			return nil, err
		}

		if n, ok := res.(*yaml.Node); ok {
			return n, nil
		}

		in = res
	}

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	//nolint:exhaustive
	switch v.Kind() {
	case reflect.Struct:
		node.Kind = yaml.MappingNode

		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			// skip unexported fields
			if !v.Field(i).CanInterface() {
				continue
			}

			comment := t.Field(i).Tag.Get("comment")
			tag := t.Field(i).Tag.Get("yaml")
			parts := strings.Split(tag, ",")
			fieldName := parts[0]
			parts = parts[1:]

			if fieldName == "" {
				fieldName = strings.ToLower(t.Field(i).Name)
			}

			if fieldName == "-" {
				continue
			}

			var (
				empty = isEmpty(v.Field(i))
				null  = isNil(v.Field(i))

				skip   bool
				inline bool
				flow   bool
			)

			for _, part := range parts {
				if part == "omitempty" && empty {
					skip = true
				}

				if part == "omitonlyifnil" && !null {
					skip = false
				}

				if part == "inline" {
					inline = true
				}

				if part == "flow" {
					flow = true
				}
			}

			var value interface{}
			if v.Field(i).CanInterface() {
				value = v.Field(i).Interface()
			}

			if skip {
				continue
			}

			var style yaml.Style
			if flow {
				style |= yaml.FlowStyle
			}

			if inline {
				child, err := toYamlNode(value)
				if err != nil {
					return nil, err
				}

				if child.Kind == yaml.MappingNode || child.Kind == yaml.SequenceNode {
					appendNodes(node, child.Content...)
				}
			} else if err := addToMap(node, comment, fieldName, value, style); err != nil {
				return nil, err
			}
		}
	case reflect.Map:
		node.Kind = yaml.MappingNode
		keys := v.MapKeys()
		// always interate keys in alphabetical order to preserve the same output for maps
		sort.Slice(
			keys, func(i, j int) bool {
				return keys[i].String() < keys[j].String()
			},
		)

		for _, k := range keys {
			element := v.MapIndex(k)
			value := element.Interface()

			if err := addToMap(node, "", k.Interface(), value, 0); err != nil {
				return nil, err
			}
		}
	case reflect.Slice:
		node.Kind = yaml.SequenceNode
		nodes := make([]*yaml.Node, v.Len())

		for i := 0; i < v.Len(); i++ {
			element := v.Index(i)

			var err error

			nodes[i], err = toYamlNode(element.Interface())
			if err != nil {
				return nil, err
			}
		}
		appendNodes(node, nodes...)
	default:
		if err := node.Encode(in); err != nil {
			return nil, err
		}
	}

	return node, nil
}

func appendNodes(dest *yaml.Node, nodes ...*yaml.Node) {
	if dest.Content == nil {
		dest.Content = []*yaml.Node{}
	}

	dest.Content = append(dest.Content, nodes...)
}

func addToMap(dest *yaml.Node, comment string, fieldName, in interface{}, style yaml.Style) error {
	key, err := toYamlNode(fieldName)
	if err != nil {
		return err
	}
	key.HeadComment = comment

	value, err := toYamlNode(in)
	if err != nil {
		return err
	}

	value.Style = style

	// override head comment with line comment for non-scalar nodes
	if value.Kind != yaml.ScalarNode {
		if key.HeadComment == "" {
			key.HeadComment = value.LineComment
		}
		value.LineComment = ""
	}

	appendNodes(dest, key, value)

	return nil
}
