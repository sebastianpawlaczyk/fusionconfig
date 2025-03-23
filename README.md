# fusionconfig

Library for reading configurations from three sources:
* environment variables
* json local file
* json remote file

## How to use it?

```go
package main

import (
	"fmt"
	"github.com/fusionconfig"
)

type AppConfiguration struct {
	Port int
}

func main() {
	cfg := AppConfiguration{}
	if err := fusionconfig.LoadConfig(&cfg,
		fusionconfig.WithEnv(true), // by default is it enabled
		fusionconfig.WithLocalFile("./fixtures/test-file.json"),
		fusionconfig.WithRemoteFile("https://some-server/json-example.file"),
	); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Port)
}
```
