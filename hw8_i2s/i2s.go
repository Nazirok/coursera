package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	dv := reflect.ValueOf(data)
	//fmt.Println(dv)
	//fmt.Println(reflect.TypeOf(data).Kind())
	switch dv.Kind() {
	case reflect.Map:
		ov := reflect.ValueOf(out)
		fmt.Println(ov)
		fmt.Println(ov.Kind())
		if ov.Kind() == reflect.Ptr {
			fmt.Println(ov.Elem().)
			//for i := 0; i < ov.NumField(); i++ {
			//	valueField := ov.Field(i)
			//	fmt.Println(valueField.String())
			//}
		}
		//for _, k := range dv.MapKeys() {
		//	fmt.Println(k, " ", dv.MapIndex(k))
		//}
	}
	return nil
}
