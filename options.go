package fusionconfig

type config[T any] struct {
	withEnv       bool
	localFile     string
	remoteUrlFile string
	prefix        string

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
		config.localFile = filePath
	}
}

func WithRemoteFile[T any](fileUrl string) Option[T] {
	return func(config *config[T]) {
		config.remoteUrlFile = fileUrl
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
