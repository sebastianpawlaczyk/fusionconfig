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

func SetInt(val reflect.Value, value string, size int) error {
	intVal, err := strconv.ParseInt(value, 10, size)
	if err != nil {
		return fmt.Errorf("failed to convert %s to int for key %s: %w", val, value, err)
	}
	val.SetInt(intVal)
	return nil
}

func SetUint(val reflect.Value, value string, size int) error {
	uintVal, err := strconv.ParseUint(value, 10, size)
	if err != nil {
		return fmt.Errorf("failed to convert %s to uint: %w", value, err)
	}
	val.SetUint(uintVal)
	return nil
}

func SetFloat(val reflect.Value, value string, size int) error {
	floatVal, err := strconv.ParseFloat(value, size)
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
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
				err = SetUint(ele, data[i], ele.Type().Bits())
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
				err = SetInt(ele, data[i], ele.Type().Bits())
			case reflect.Float32, reflect.Float64:
				err = SetFloat(ele, data[i], ele.Type().Bits())
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
