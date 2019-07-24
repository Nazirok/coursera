package main

import (
	"errors"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	vd := reflect.ValueOf(data)
	vo := reflect.ValueOf(out)
	if vo.Kind() == reflect.Ptr {
		vo = vo.Elem()
	}
	if err := set(vd, vo); err != nil {
		return err
	}
	return nil
}


func set(mv, sv reflect.Value) error {
	if !sv.IsValid() {
		return errors.New("no such field")
	}
	if !sv.CanSet() {
		return errors.New("can`t set field")
	}
	switch sv.Kind() {
	case reflect.Int:
		switch mv.Kind() {
		case reflect.Int:
			sv.SetInt(mv.Int())
		case reflect.Float64:
			sv.SetInt(int64(mv.Float()))
		}
	case reflect.String:
		sv.SetString(mv.String())
	case reflect.Bool:
		sv.SetBool(mv.Bool())
	case reflect.Slice:
		fmt.Println(mv.String())
	case reflect.Struct:
		switch mv.Kind() {
		case reflect.Map:
			for _, key := range mv.MapKeys() {
				ssv := sv.FieldByName(key.String())
				mmv := mv.MapIndex(key).Elem()
				if err := set(mmv, ssv); err != nil {
					return err
				}
			}
		}
		fmt.Println("sdfsdf",sv.String())
		fmt.Println(sv.CanSet())
	default:
	}
	return nil
}
