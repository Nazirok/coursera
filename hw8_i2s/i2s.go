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
		fmt.Println(vo.Kind())
	}
	switch vd.Kind() {
	case reflect.Map:
		for _, key := range vd.MapKeys() {
			fo := vo.FieldByName(key.String())
			vdKeyV := vd.MapIndex(key).Elem()
			if err := setStructField(vdKeyV, fo); err != nil {
				return err
			}
		}

	}
	return nil
}

func setStructField(mv, sv reflect.Value) error {
	if !sv.IsValid() {
		return errors.New("no such field")
	}
	if !sv.CanSet() {
		return errors.New("can`t set field")
	}
	switch mv.Kind() {
	case reflect.Float64:
		if sv.Kind() == reflect.Int {
			sv.SetInt(int64(mv.Float()))
		}
	case reflect.String:
		if sv.Kind() == reflect.String {
			sv.SetString(mv.String())
		}
	case reflect.Bool:
		if sv.Kind() == reflect.Bool {
			sv.SetBool(mv.Bool())
		}

	}
	return nil
}
