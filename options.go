package fusionconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type config[T any] struct {
	prefix  string
	withEnv bool

	sources     []func(keys []string) (map[string]string, error)
	validations []func(obj T) error
}

type Option[T any] func(*config[T])

func WithEnv[T any](withEnv bool) Option[T] {
	return func(config *config[T]) {
		config.withEnv = withEnv
	}
}

func WithLocalFile[T any](filePath string) Option[T] {
	return func(config *config[T]) {
		config.sources = append(config.sources, func(keys []string) (map[string]string, error) {
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
		})
	}
}

func WithRemoteFile[T any](fileUrl string) Option[T] {
	return func(config *config[T]) {
		config.sources = append(config.sources, func(keys []string) (map[string]string, error) {
			client := http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(fileUrl)
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
		})
	}
}

func WithPrefix[T any](prefix string) Option[T] {
	return func(config *config[T]) {
		config.prefix = prefix
	}
}

func WithValidation[T any](fn func(T) error) Option[T] {
	return func(cfg *config[T]) {
		cfg.validations = append(cfg.validations, fn)
	}
}
