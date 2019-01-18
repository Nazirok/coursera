package main

import (
	"reflect"
	"fmt"
)

func i2s(data interface{}, out interface{}) error {
	v := reflect.ValueOf(data)
	fmt.Println(reflect.TypeOf(data).Kind())
	switch reflect.TypeOf(data).Kind() {
	case reflect.Map:
		fmt.Println("Struct")
	}
	return nil
}

