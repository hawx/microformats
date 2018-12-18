// Package microformats allows parsing of microformats2 data into structs.
//
// This package is a wrapper around https://github.com/andyleap/microformats to
// replicate the structtag based parsing of encoding/json (and others).
//
// Note that, contrary to encoding/json, fields must be tagged with
// `mf:"PROPERTY"` to be set.
package microformats

import (
	"errors"
	"io"
	"net/url"
	"reflect"

	mf2 "github.com/andyleap/microformats"
)

// TODO: Handle pointer fields

// ParseAll will fill a slice with instances of the tag found when parsing r.
func ParseAll(r io.Reader, baseURL *url.URL, tag string, v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a pointer to a non-nil struct")
	}
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	ty := rv.Type().Elem()
	keyMap := makeKeyMap(ty)

	parser := mf2.New()
	data := parser.Parse(r, baseURL)

	for _, item := range data.Items {
		if contains(tag, item.Type) {
			inst := reflect.New(ty).Elem()
			inst = parseWithKeyMap(inst, keyMap, item)

			rv.Set(reflect.Append(rv, inst))
		}
	}

	return nil
}

// Parse will find within r any instances of the microformat given by tag,
// resolving any URLs encountered against the baseURL provded, and assign the
// tagged properties to v.
func Parse(r io.Reader, baseURL *url.URL, tag string, v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a pointer to a non-nil struct")
	}
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	keyMap := makeKeyMap(rv.Type())

	parser := mf2.New()
	data := parser.Parse(r, baseURL)

	for _, item := range data.Items {
		if contains(tag, item.Type) {
			rv = parseWithKeyMap(rv, keyMap, item)
		}
	}

	return nil
}

func parse(rv reflect.Value, obj interface{}) reflect.Value {
	keyMap := makeKeyMap(rv.Type())

	return parseWithKeyMap(rv, keyMap, obj)
}

func parseWithKeyMap(rv reflect.Value, keyMap map[string]reflect.StructField, obj interface{}) reflect.Value {
	item := obj.(*mf2.MicroFormat)

	for key, value := range item.Properties {
		field := getField(rv, keyMap, key)
		trySet(field, value)
	}

	return rv
}

func contains(needle string, xs []string) bool {
	for _, x := range xs {
		if x == needle {
			return true
		}
	}

	return false
}

func makeKeyMap(rt reflect.Type) map[string]reflect.StructField {
	keyMap := map[string]reflect.StructField{}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if mf, ok := field.Tag.Lookup("mf"); ok && mf != "" {
			keyMap[mf] = field
		}
	}

	return keyMap
}

func getField(rv reflect.Value, keyMap map[string]reflect.StructField, key string) reflect.Value {
	return rv.FieldByName(keyMap[key].Name)
}

func trySet(field reflect.Value, values []interface{}) {
	if !field.IsValid() {
		return
	}

	switch field.Type().Kind() {
	case reflect.Slice:
		for _, value := range values {
			field.Set(reflect.Append(field, reflect.ValueOf(value)))
		}
	case reflect.String:
		field.SetString(values[0].(string))
	case reflect.Struct:
		field.Set(parse(field, values[0]))
	}
}
