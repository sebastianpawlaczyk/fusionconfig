package fusionconfig

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"os"
	"reflect"
	"strings"

	"github.com/sp/fusionconfig/utils"
)

const (
	tag = "fusionconfig"
)

type Validation[T any] interface {
	Validate(in T) error
}

func LoadConfig[T any](obj T, opt ...Option[T]) error {
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("obj must be a struct")
	}

	cfg := config[T]{
		withEnv: true,
	}
	for _, o := range opt {
		o(&cfg)
	}

	if v, ok := any(obj).(Validation[T]); ok {
		cfg.validations = append(cfg.validations, v.Validate)
	}

	keys := getKeys(elem, cfg.prefix)

	if cfg.withEnv {
		fn := func(keys []string) (map[string]string, error) {
			return getFromEvn(keys), nil
		}

		cfg.sources = append(
			[]func(keys []string) (map[string]string, error){fn},
			cfg.sources...,
		)
	}

	results := make([]map[string]string, len(cfg.sources))

	errGroup := errgroup.Group{}
	for i, fn := range cfg.sources {
		errGroup.Go(func() error {
			v, err := fn(keys)
			results[i] = v

			return err
		})
	}

	if err := errGroup.Wait(); err != nil {
		return err
	}

	merge := mergeMaps(results)

	if err := populateFields(elem, merge, cfg.prefix); err != nil {
		return err
	}

	for _, v := range cfg.validations {
		if err := v(obj); err != nil {
			return err
		}
	}

	return nil
}

func flattenKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func getKeys(val reflect.Value, prefix string) []string {
	keys := make([]string, 0, val.NumField())

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)
		if !fieldVal.CanInterface() {
			continue
		}

		key := field.Tag.Get(tag)
		if key == "" {
			key = field.Name
		}

		fullKey := flattenKey(prefix, key)
		if fieldVal.Kind() == reflect.Struct {
			nestedVars := getKeys(fieldVal, fullKey)
			keys = append(keys, nestedVars...)
		} else {
			keys = append(keys, fullKey)
		}
	}

	return keys
}

func getFromEvn(keys []string) map[string]string {
	res := make(map[string]string)
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		if !ok {
			continue
		}

		res[k] = v
	}

	return res
}

func flattenMap(prefix string, data map[string]interface{}, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			flattenMap(fullKey, v, result)
		case []interface{}:
			var parts []string
			for _, item := range v {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
			result[fullKey] = strings.Join(parts, ",")
		default:
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}

func mergeMaps(in []map[string]string) map[string]string {
	result := make(map[string]string)

	for _, m := range in {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

func populateFields(val reflect.Value, configMap map[string]string, prefix string) error {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		key := field.Tag.Get(tag)
		if key == "" {
			key = field.Name
		}
		fullKey := flattenKey(prefix, key)

		if fieldVal.Kind() == reflect.Struct {
			if err := populateFields(fieldVal, configMap, fullKey); err != nil {
				return err
			}
			continue
		}

		if value, ok := configMap[fullKey]; ok {
			var err error
			switch fieldVal.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Bool:
				err = utils.SetBoolValue(fieldVal, value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				err = utils.SetInt(fieldVal, value, field.Type.Bits())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				err = utils.SetUint(fieldVal, value, field.Type.Bits())
			case reflect.Float32, reflect.Float64:
				err = utils.SetFloat(fieldVal, value, field.Type.Bits())
			case reflect.Slice:
				err = utils.SetValueWithSlice(fieldVal, value, ",")
			default:
				return fmt.Errorf("unsupported field type: %s", fieldVal.Kind())
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}
