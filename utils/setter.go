package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func SetBoolValue(v reflect.Value, value string) error {
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}

	v.SetBool(boolValue)
	return nil
}

func SetInt(val reflect.Value, value string) error {
	intVal, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert %s to int for key %s: %w", val, value, err)
	}
	val.SetInt(intVal)
	return nil
}

func SetUint(val reflect.Value, value string) error {
	uintVal, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to convert %s to uint: %w", value, err)
	}
	val.SetUint(uintVal)
	return nil
}

func SetFloat(val reflect.Value, value string) error {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("failed to convert %s to float: %w", value, err)
	}
	val.SetFloat(floatVal)
	return nil
}

func SetValueWithSlice(v reflect.Value, slice string, separator string) error {
	data := strings.Split(slice, separator)
	size := len(data)
	if size > 0 {
		slice := reflect.MakeSlice(v.Type(), size, size)
		for i := 0; i < size; i++ {
			ele := slice.Index(i)
			kind := ele.Kind()
			var err error
			switch kind {
			case reflect.Bool:
				err = SetBoolValue(ele, data[i])
			case reflect.String:
				ele.SetString(data[i])
			case reflect.Uint8:
				err = SetUint(ele, data[i])
			case reflect.Uint16:
				err = SetUint(ele, data[i])
			case reflect.Uint, reflect.Uint32:
				err = SetUint(ele, data[i])
			case reflect.Uint64:
				err = SetUint(ele, data[i])
			case reflect.Int8:
				err = SetInt(ele, data[i])
			case reflect.Int16:
				err = SetInt(ele, data[i])
			case reflect.Int, reflect.Int32:
				err = SetInt(ele, data[i])
			case reflect.Int64:
				err = SetInt(ele, data[i])
			case reflect.Float32:
				err = SetFloat(ele, data[i])
			case reflect.Float64:
				err = SetFloat(ele, data[i])
			default:
				return fmt.Errorf("Can't support type: %s", kind.String())
			}

			if err != nil {
				return err
			}
		}

		v.Set(slice)
	}

	return nil
}
