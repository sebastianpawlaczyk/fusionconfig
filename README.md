# fusionconfig

Library for reading configurations from three sources:
* environment variables
* json local file
* json remote file

**Reading Priority:** remote file, local file, env variables

## How to use it?

```go
package main

import (
	"errors"
	"fmt"
	"github.com/fusionconfig"
)

type AppConfiguration struct {
	Port string
}

func main() {
	validationFunc := func(obj any) error {
		cfg, ok := obj.(*AppConfiguration)
		if !ok {
			return errors.New("invalid configuration type")
		}

		if cfg.Port == "" {
			return errors.New("port cannot be empty")
		}

		return nil
	}

	cfg := AppConfiguration{}
	if err := fusionconfig.LoadConfig(&cfg,
		fusionconfig.WithEnv(true), // by default is it enabled
		fusionconfig.WithLocalFile("./fixtures/test-file.json"),
		fusionconfig.WithRemoteFile("https://some-server/json-example.file"),
		fusionconfig.WithValidation(validationFunc),
	); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Port)
}
```
