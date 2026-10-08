package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// object is a JSON object that keeps its keys in the order they were read, so
// that writing a document back changes nothing but what was meant to change.
type object struct {
	keys   []string
	values map[string]any
}

func newObject() *object {
	return &object{values: map[string]any{}}
}

func (o *object) get(key string) any {
	return o.values[key]
}

func (o *object) getObject(key string) *object {
	child, _ := o.values[key].(*object)
	return child
}

// set replaces a key's value, or adds the key where it sorts among the others:
// the converter writes a schema's keys in order.
func (o *object) set(key string, value any) {
	if _, found := o.values[key]; !found {
		at := len(o.keys)
		for i, existing := range o.keys {
			if existing > key {
				at = i
				break
			}
		}
		o.keys = slices.Insert(o.keys, at, key)
	}
	o.values[key] = value
}

func (o *object) remove(key string) {
	if _, found := o.values[key]; !found {
		return
	}
	delete(o.values, key)
	o.keys = slices.DeleteFunc(o.keys, func(k string) bool { return k == key })
}

// decodeDocument reads JSON keeping the order of every object's keys and every
// number as it was written.
func decodeDocument(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("more than one JSON value")
	}
	return value, nil
}

func decodeValue(decoder *json.Decoder) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}
	switch delim {
	case '{':
		obj := newObject()
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, _ := keyToken.(string)
			value, err := decodeValue(decoder)
			if err != nil {
				return nil, err
			}
			if _, dup := obj.values[key]; !dup {
				obj.keys = append(obj.keys, key)
			}
			obj.values[key] = value
		}
		_, err = decoder.Token()
		return obj, err
	case '[':
		list := []any{}
		for decoder.More() {
			value, err := decodeValue(decoder)
			if err != nil {
				return nil, err
			}
			list = append(list, value)
		}
		_, err = decoder.Token()
		return list, err
	default:
		return nil, fmt.Errorf("unexpected %v", delim)
	}
}

// encodeDocument writes JSON the way the OpenAPI converter does - Jackson's
// default pretty printer: two spaces an object level, `"key" : value`, arrays
// on the line they open on and not indented further, `{ }` and `[ ]` for empty
// ones, non-ASCII as it is, and no newline at the end.
func encodeDocument(value any) []byte {
	var buf bytes.Buffer
	writeValue(&buf, value, 0)
	return buf.Bytes()
}

func writeValue(buf *bytes.Buffer, value any, level int) {
	switch v := value.(type) {
	case *object:
		if len(v.keys) == 0 {
			buf.WriteString("{ }")
			return
		}
		buf.WriteString("{")
		for i, key := range v.keys {
			if i > 0 {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
			buf.WriteString(strings.Repeat("  ", level+1))
			writeString(buf, key)
			buf.WriteString(" : ")
			writeValue(buf, v.values[key], level+1)
		}
		buf.WriteString("\n")
		buf.WriteString(strings.Repeat("  ", level))
		buf.WriteString("}")
	case []any:
		if len(v) == 0 {
			buf.WriteString("[ ]")
			return
		}
		buf.WriteString("[ ")
		for i, item := range v {
			if i > 0 {
				buf.WriteString(", ")
			}
			writeValue(buf, item, level)
		}
		buf.WriteString(" ]")
	case string:
		writeString(buf, v)
	case json.Number:
		buf.WriteString(v.String())
	case bool:
		if v {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case nil:
		buf.WriteString("null")
	default:
		panic(fmt.Sprintf("openapi: cannot write %T", value))
	}
}

// writeString escapes what JSON requires and nothing more, as Jackson does.
func writeString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		default:
			if r < 0x20 {
				fmt.Fprintf(buf, `\u%04X`, r)
			} else {
				buf.WriteRune(r)
			}
		}
	}
	buf.WriteByte('"')
}
