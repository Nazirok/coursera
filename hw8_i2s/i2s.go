package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	vd := reflect.ValueOf(data)
	vo := reflect.ValueOf(out)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	switch vd.Kind() {
	case reflect.Map:
		for _, key := range vd.MapKeys() {
			fo := vo.FieldByName(key.String())
			if fo.IsValid() {
				if fo.CanSet() {
					vdKeyV := vd.MapIndex(key).Elem()
					switch vdKeyV.Kind() {
					case reflect.Float64:
						if fo.Kind() == reflect.Int {
							fo.SetInt(int64(vdKeyV.Float()))
						}
					case reflect.String:
						if fo.Kind() == reflect.String {
							fo.SetString(vdKeyV.String())
						}
					case reflect.Bool:
						if fo.Kind() == reflect.Bool {
							fo.SetBool(vdKeyV.Bool())
						}
					case reflect.Slice:
						fmt.Println(vdKeyV.String())
					case reflect.Struct:
						fmt.Println(vdKeyV.String())
					default:
						fmt.Println(vdKeyV.Kind())
					}
				}
			}

		}

	}
	return nil
}
