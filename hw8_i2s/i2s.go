package main

import (
	"errors"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	vd := reflect.ValueOf(data)
	vo := reflect.ValueOf(out)
	if vo.Kind() != reflect.Ptr {
		return errors.New("out type must be ptr")
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
		default:
			return errors.New(fmt.Sprintf("type %s can`t set to %s", mv.Kind(), sv.Kind()))
		}
	case reflect.String:
		if mv.Kind() != reflect.String {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", mv.Kind(), sv.Kind()))
		}
		sv.SetString(mv.String())
	case reflect.Bool:
		if mv.Kind() != reflect.Bool {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", mv.Kind(), sv.Kind()))
		}
		sv.SetBool(mv.Bool())
	case reflect.Slice:
		if mv.Kind() != reflect.Slice {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", mv.Kind(), sv.Kind()))
		}
		slice := reflect.MakeSlice(sv.Type(), mv.Len(), mv.Len())
		for i := 0; i < mv.Len(); i++ {
			if err := set(mv.Index(i).Elem(), slice.Index(i)); err != nil {
				return err
			}
		}
		sv.Set(slice)
	case reflect.Struct:
		if mv.Kind() != reflect.Map {
			return errors.New(fmt.Sprintf("type %s can`t set to %s", mv.Kind(), sv.Kind()))
		}
		for _, key := range mv.MapKeys() {
			ssv := sv.FieldByName(key.String())
			mmv := mv.MapIndex(key).Elem()
			if err := set(mmv, ssv); err != nil {
				return err
			}
		}
	default:
		return errors.New("unknown error")
	}
	return nil
}

