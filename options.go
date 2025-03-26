package fusionconfig

import "fmt"

type config struct {
	withEnv       bool
	localFile     string
	remoteUrlFile string
	prefix        string

	validations []func(obj any) error
}

type Option func(*config)

func WithEnv(withEnv bool) Option {
	return func(config *config) {
		config.withEnv = withEnv
	}
}

func WithLocalFile(filePath string) Option {
	return func(config *config) {
		config.localFile = filePath
	}
}

func WithRemoteFile(fileUrl string) Option {
	return func(config *config) {
		config.remoteUrlFile = fileUrl
	}
}

func WithPrefix(prefix string) Option {
	return func(config *config) {
		config.prefix = prefix
	}
}

func WithValidation[T any](fn func(*T) error) Option {
	return func(cfg *config) {
		cfg.validations = append(cfg.validations, func(obj any) error {
			t, ok := obj.(*T)
			if !ok {
				return fmt.Errorf("invalid type passed to validation: expected %T", *new(T))
			}
			return fn(t)
		})
	}
}
