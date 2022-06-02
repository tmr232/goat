# goat 🐐
GO Approximation of Typer 


## Status

This is still in experimental state, API is unstable and may break.


## Example

Try it yourself!

```go
package main

import (
	"fmt"
	"github.com/tmr232/goat"
)

//go:generate go run github.com/tmr232/goat/cmd/goat

func app(name string, goodbye bool) {
	if goodbye {
		fmt.Printf("Goodbye, %s.\n", name)
	} else {
		fmt.Printf("Hello, %s!\n", name)

	}
}

func main() {
	goat.Run(app)
}

```

Note that currently, only `string` and `bool` arguments are supported.

Additionally, you'll need the [urfave/cli](https://github.com/urfave/cli) package to run the resulting app. 


See [this blog post](https://blog.tamir.dev/posts/goat-codegen-initial/) for more implementation details.