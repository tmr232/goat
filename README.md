# goat 🐐
GO Approximation of [Typer][Typer]

## Intro

[Typer][Typer] is one of my favourite Python packages.
It allows building CLI apps with next to zero boilerplate.
See [the Typer docs](https://typer.tiangolo.com/#the-absolute-minimum) for an example.

Working in Go, I want something similar.
So I am creating Goat 🐐, trying to get the best experience possible.

```go
package main

import (
	"fmt"
	"github.com/tmr232/goat" // One explicit dependency
)

// Generate all the necessary magic
//go:generate go run github.com/tmr232/goat/cmd/goater

func app(name string, goodbye bool) {
	if goodbye {
		fmt.Printf("Goodbye, %s.\n", name)
	} else {
		fmt.Printf("Hello, %s!\n", name)

	}
}

func main() {
	// Let goat know what to run
	goat.Run(app)
}

```

## Status & Contibuting

Slowly moving forward, but not yet stable.

Experimentation, bug reports, and feature requests are very welcome.

## API

The Goat API is built of 4 main parts, numbered in the code below.

```go
package main

import (
	"fmt"
	"github.com/tmr232/goat"
)
// (1) The goater command
//go:generate go run github.com/tmr232/goat/cmd/goater

// (3) The app function signature
func app(name string, goodbye bool, question *string, times int) {
	// (4) Flag Descriptors
	goat.Flag(name).
		Usage("The name to greet")
	goat.Flag(goodbye).
		Name("bye").
		Usage("Enable to say Goodbye")
	goat.Flag(question).
		Usage("Instead of a greeting, ask a question.")
	goat.Flag(times).
		Usage("Number of repetitions").
		Default(1)

	for i := 0; i < times; i++ {
		if question != nil {
			fmt.Printf("%s, %s?", *question, name)
		} else {
			if goodbye {
				fmt.Printf("Goodbye, %s.\n", name)
			} else {
				fmt.Printf("Hello, %s!\n", name)
			}
		}
	}
}

func main() {
	// (2) The Run function
	goat.Run(app)
}

```

### 1. The `goater` Command

To get the functionality we aim for, we use code generation. 
We read the user's code, infer the relevant information, 
and generate wrapper code that later calls into it.
This generation is done using the `goater` command.
You can either run it manually in the relevant package directory, 
or use `go:generate` to do it for you.

### 2. The `goat.Run` Function

To let the `goater` command know which functions to wrap, 
we call `goat.Run` with those function. 
In this case - `goat.Run(app)` means that we'll wrap the `app` function.

Later, during execution, `goat.Run` calls into the wrapper for `app`.

### 3. The App Function Signature

This is where things get interesting.
The app function (any function passed into `goat.Run`) is parsed during code
generation, and a wrapper is generated for it.

The signature of the function determines the types of flags that will be generated.
An `int` argument will result in an `int` flag, a `bool` argument in a `bool` flag, 
and a `string` argument in a `string` flag.

If an argument is a pointer (`*string`, for example) the flag will be optional.
If an argument is not a pointer, it'll be a required flag.
`bool` is an exception as it is never required.

### 4. Flag Descriptors

A name and a type for a flag are nice, but hardly enough.
We may want to define aliases, usage strings, or default values.
To do this - we describe our flags as follows:

```go
goat.Flag(name).
	Usage("The name to greet")
```

You can use the following to add data to your flags:

1. `Usage(string)` - add a usage string
2. `Name(string)` - set the name of the flag
3. `Default(any)` - set the flag's default value. Works only with non-pointer flags.

## Subcommands & Context

Goat also allows defining subcommands

```go
package main

import (
	"fmt"
	"github.com/tmr232/goat" // One explicit dependency
)

//go:generate go run github.com/tmr232/goat/cmd/goater

func server(name string) {
	
}

func app(name string, goodbye bool) {
	if goodbye {
		fmt.Printf("Goodbye, %s.\n", name)
	} else {
		fmt.Printf("Hello, %s!\n", name)

	}
}

func main() {
	// Let goat know what to run
	goat.Run(app)
}

```

## Dependencies

Goat currently uses [urfave/cli](https://github.com/urfave/cli) for parsing flags.
Other than that, the generated code currently only depends on the standard library.

In the future, I plan to write backends for other popular flag-parsing libraries
(namely [Cobra](https://cobra.dev/) and [flag](https://pkg.go.dev/flag)) so that
users can choose what they depend on.


[Typer]:https://typer.tiangolo.com/
