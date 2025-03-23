package fusionconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fusionconfig/utils"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	tag = "fusionconfig"
)

func LoadConfig(obj any, opt ...Option) error {
	cfg := config{
		withEnv:       true,
		localFile:     "",
		remoteUrlFile: "",
		prefix:        "",
	}
	for _, o := range opt {
		o(&cfg)
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("obj must be a struct")
	}

	keys := getKeys(elem, cfg.prefix)

	resChan := make(chan map[string]string, 1)
	localFileChan := make(chan map[string]string, 1)
	remoteFileChan := make(chan map[string]string, 1)
	errChan := make(chan error, 2)

	wg := sync.WaitGroup{}

	if cfg.withEnv {
		wg.Add(1)

		go func() {
			defer wg.Done()
			getFromEvn(keys, resChan)
		}()
	}

	if cfg.localFile != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()
			res, err := getFromFile(cfg.localFile, keys)
			if err != nil {
				errChan <- err
			} else {
				localFileChan <- res
			}
		}()
	}

	if cfg.remoteUrlFile != "" {
		wg.Add(1)

		go func() {
			defer wg.Done()
			res, err := getFromRemoteFile(cfg.remoteUrlFile, keys)
			if err != nil {
				errChan <- err
			} else {
				remoteFileChan <- res
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resChan)
		close(localFileChan)
		close(remoteFileChan)
		close(errChan)
	}()

	for err := range errChan {
		return err
	}

	envMap := <-resChan
	localFileMap := <-localFileChan
	remoteFileMap := <-remoteFileChan

	merge := mergeMaps(envMap, localFileMap, remoteFileMap)

	return populateFields(elem, merge, cfg.prefix)
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

func getFromEvn(keys []string, r chan map[string]string) {
	res := make(map[string]string)
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		if !ok {
			continue
		}

		res[k] = v
	}

	r <- res
}

func getFromFile(filePath string, keys []string) (map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	var rawData map[string]interface{}
	if err = json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	flattened := make(map[string]string)
	flattenMap("", rawData, flattened)

	result := make(map[string]string)
	for _, key := range keys {
		if value, ok := flattened[key]; ok {
			result[key] = value
		}
	}

	return result, nil
}

func getFromRemoteFile(url string, keys []string) (map[string]string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching remote file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch remote config error, reply status code is %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading remote response body: %w", err)
	}

	var rawData map[string]interface{}
	if err = json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("error unmarshaling remote JSON: %w", err)
	}

	flattened := make(map[string]string)
	flattenMap("", rawData, flattened)

	result := make(map[string]string)
	for _, key := range keys {
		if value, ok := flattened[key]; ok {
			result[key] = value
		}
	}

	return result, nil
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

func mergeMaps(base, override, override2 map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	for k, v := range override2 {
		result[k] = v
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
			switch fieldVal.Kind() {
			case reflect.String:
				fieldVal.SetString(value)
			case reflect.Bool:
				err := utils.SetBoolValue(fieldVal, value)
				if err != nil {
					return err
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				err := utils.SetInt(fieldVal, value)
				if err != nil {
					return err
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				err := utils.SetUint(fieldVal, value)
				if err != nil {
					return err
				}
			case reflect.Float32, reflect.Float64:
				err := utils.SetFloat(fieldVal, value)
				if err != nil {
					return err
				}
			case reflect.Slice:
				err := utils.SetValueWithSlice(fieldVal, value, ",")
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported field type: %s", fieldVal.Kind())
			}
		}
	}
	return nil
}
