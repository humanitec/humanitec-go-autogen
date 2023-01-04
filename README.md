# humanitec-go-autogen

Autogenerated humanitec golang client

Usage

```golang
package cmd

import (
	"github.com/humanitec/humanitec-go-autogen"
)

func doSomething() {
	client, err := humanitec.NewClient(&humanitec.Config{
		Token: os.Getenv("HUM_TOKEN"),
	})
}
```
