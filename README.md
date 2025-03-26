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
	"fmt"

	"github.com/sp/fusionconfig"
)

type AppConfiguration struct {
	Port string
}

func (a AppConfiguration) Validate() error {
	if len(a.Port) == 0 {
		return fmt.Errorf("port is required")
	}
	return nil
}

func main() {
	cfg := AppConfiguration{}
	if err := fusionconfig.LoadConfig(&cfg,
		fusionconfig.WithEnv(true), // by default is it enabled
		fusionconfig.WithLocalFile("./fixtures/test-file.json"),
		fusionconfig.WithRemoteFile("https://some-server/json-example.file"),
		fusionconfig.WithValidation(func(t *AppConfiguration) error {
			if len(t.Port) == 0 {
				return fmt.Errorf("port is required")
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Port)
}
```
