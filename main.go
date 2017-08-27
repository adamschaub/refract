package main

import (
	"fmt"
	"reflect"
	"strings"
)

type Test struct {
	Field int      `desc:"an int field"`
	Sub   string   `desc:"foo foo foo"`
	List  []string `desc:"slice of strings"`
}

func main() {
	printFieldsAndDesc(&Test{})
	return
}

func printFieldsAndDesc(gCfg interface{}) error {
	return eachSubField(gCfg, func(parent reflect.Value, subFieldName string, crumbs []string) error {
		p := strings.Join(crumbs, ".")
		subField, _ := parent.Type().FieldByName(subFieldName)
		desc := subField.Tag.Get("desc")

		fmt.Printf("Field: %s.%s (%s) -> ", p, subFieldName, desc)
		switch subField.Type.Kind() {
		case reflect.Bool:
			fmt.Println("Bool")
		case reflect.Int:
			fmt.Println("Int")
		case reflect.Int64:
			fmt.Println("Int64")
		case reflect.String:
			fmt.Println("String")
		case reflect.Float64:
			fmt.Println("Float64")
		case reflect.Slice:
			fmt.Println("Slice")
		default:
			return fmt.Errorf("%s is unsupported by config @ %s.%s", subField.Type.String(), p, subFieldName)
		}
		return nil
	})
}

// eachSubField is used for a struct of structs (like GlobalConfig). fn is called
// with each field from each sub-struct of the parent. Fields are skipped if they
// are not settable, or unexported OR are marked with `flag:"false"`
func eachSubField(i interface{}, fn func(reflect.Value, string, []string) error, crumbs ...string) error {
	t := reflect.ValueOf(i)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic("eachSubField can only be called on a pointer-to-struct")
	}
	// Sanity check. Should be true if it is a pointer-to-struct
	if !t.Elem().CanSet() {
		panic("eachSubField can only be called on a settable struct of structs")
	}

	t = t.Elem()
	nf := t.NumField()
	for i := 0; i < nf; i++ {
		field := t.Field(i)
		sf := t.Type().Field(i)
		if sf.Tag.Get("flag") == "false" {
			continue
		}

		if field.Kind() == reflect.Struct && field.CanSet() {
			eachSubField(field.Addr().Interface(), fn, append(crumbs, sf.Name)...)
		} else if field.CanSet() {
			if err := fn(t, sf.Name, crumbs); err != nil {
				return err
			}
		}
	}
	return nil
}
