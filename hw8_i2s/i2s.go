package main

import (
	"errors"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	vd := reflect.ValueOf(data)
	vo := reflect.ValueOf(out)
	if err := set(vd, vo); err != nil {
		return err
	}
	return nil
}

func set(mv, sv reflect.Value) error {
	if !sv.IsValid() {
		return errors.New("no such field")
	}
	switch sv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if err := set(mv, sv.Elem()); err != nil {
			return err
		}
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
		slice := reflect.MakeSlice(sv.Type(), mv.Len(), mv.Len())
		if mv.Kind() == reflect.Slice {
			for i := 0; i < mv.Len(); i++ {
				if err := set(mv.Index(i).Elem(), slice.Index(i)); err != nil {
					return err
				}
			}
		}
		sv.Set(slice)
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
	default:
	}
	return nil
}
