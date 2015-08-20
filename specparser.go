package main

import (
	"fmt"
	"github.com/opencontainers/specs"
	"reflect"
	"strings"
)

// A field represents a single field found in a struct.
type Field struct {
	Name   string
	Fields []Field //recusive
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func GetFiledJsonName(parent string, t reflect.Type) []Field {
	var fields []Field
	if t.Kind() != reflect.Struct {
		panic("GetFiledName: Type should be struct")
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag := f.Tag.Get("json")
		var name string
		if tag != "" && tag != "-" {
			name, _ = parseTag(tag)
			if parent != "" {
				name = parent + "." + name
			}
			f := Field{Name: name}
			fields = append(fields, f)
		}

		switch f.Type.Kind() {
		case reflect.Struct:
			subFs := GetFiledJsonName(name, f.Type)
			for _, subF := range subFs {
				fields = append(fields, subF)
			}
		case reflect.Slice:
			ft := f.Type.Elem()
			switch ft.Kind() {
			case reflect.Struct:
				subFs := GetFiledJsonName(name, ft)
				for _, subF := range subFs {
					fields = append(fields, subF)
				}
			case reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Func, reflect.Chan, reflect.Interface, reflect.Ptr, reflect.UnsafePointer:
			default:
			}
		//case reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Func, reflect.Chan, reflect.Interface, reflect.Ptr, reflect.UnsafePointer:
		default:
		}
	}
	return fields
}

/**
*
 */
func Print(fields []Field) {
	for _, f := range fields {
		fmt.Println(f.Name)
		Print(f.Fields)
	}
}

/**
*compare the spec's struct and the runc's struct
 */
var spport []string
var unspport []string

func CompareFields(spec []Field, runc []Field) {
	for _, fs := range spec {
		var match bool = false
		for _, fr := range runc {
			if fs.Name == fr.Name {
				match = true
				spport = append(spport, fs.Name)
				CompareFields(fs.Fields, fr.Fields)
				break
			}
		}
		if !match {
			unspport = append(unspport, fs.Name)
		}
	}
}

func main() {
	ls := specs.LinuxSpec{}
	t := reflect.ValueOf(&ls).Elem().Type()
	fmt.Printf("Number of fields:%d\n", t.NumField())
	f := GetFiledJsonName("", t)
	Print(f)

	var spec []Field = []Field{
		{
			Name: "a",
			Fields: []Field{
				{
					Name: "a1",
					Fields: []Field{
						{
							Name: "a11",
						},
					},
				},
			},
		},
		{
			Name: "b",
			Fields: []Field{
				{
					Name: "b1",
					Fields: []Field{
						{
							Name: "b11",
						},
					},
				},
			},
		},
	}

	var runc []Field = []Field{
		{
			Name: "a",
			Fields: []Field{
				{
					Name: "a1",
					Fields: []Field{
						{
							Name: "a22",
						},
					},
				},
			},
		},
		{
			Name: "c",
			Fields: []Field{
				{
					Name: "c1",
					Fields: []Field{
						{
							Name: "c11",
						},
					},
				},
			},
		},
	}

	CompareFields(spec, runc)
	fmt.Printf("=======================================\n")
	fmt.Printf("support: %v\n", spport)
	fmt.Printf("unsupport: %v\n", unspport)

}
