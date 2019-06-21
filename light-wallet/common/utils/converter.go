package utils

import (
	"reflect"
	"strings"
	"strconv"
	"fmt"
)

func Convert(Map interface{}, pointer interface{}) {
	pointertype := reflect.TypeOf(pointer)
	pointervalue := reflect.ValueOf(pointer)
	structType := pointertype.Elem()
	m := Map.(map[string]interface{})
	for i := 0; i < structType.NumField(); i++ {
		f := pointervalue.Elem().Field(i)
		stf := structType.Field(i)
		name := strings.Split(stf.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = stf.Name
		}
		v, ok := m[name]
		if !ok {
			continue
		}
		kind := pointervalue.Elem().Field(i).Kind()
		if kind == reflect.Ptr {
			kind = f.Type().Elem().Kind()
		}
		switch kind {
		case reflect.Int:
			res, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64)
			pointervalue.Elem().Field(i).SetInt(res)
			break
		case reflect.Float64:
			pointervalue.Elem().Field(i).SetFloat(v.(float64))
			break
		case reflect.Int64:
			pointervalue.Elem().Field(i).SetInt(v.(int64))
			break
		case reflect.String:
			pointervalue.Elem().Field(i).SetString(fmt.Sprint(v))
			break
		}
	}
}
