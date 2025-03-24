package fusionconfig

type config struct {
	withEnv       bool
	localFile     string
	remoteUrlFile string
	prefix        string

	validation func(obj any) error
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

func WithValidation(validation func(obj any) error) Option {
	return func(config *config) {
		config.validation = validation
	}
}
