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

## Status

This is still in experimental state, API is unstable and may break.

Additionally, the code is an undocumented mess. Be warned.


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
	goat.Describe(name).As(goat.RequiredStringFlag{Usage: "The name to greet"})
	goat.Describe(goodbye).As(goat.BoolFlag{Usage: "Enable to say Goodbye", Name: "bye"})
	goat.Describe(question).As(goat.OptionalStringFlag{Usage: "Instead of a greeting, ask a question."})
	goat.Describe(times).As(goat.DefaultIntFlag{Usage: "Number of repetitions", Default: 1})

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
To do this - we "describe" our flags as follows:

```go
//             A                    B            C
goat.Describe(name).As(goat.RequiredStringFlag{Usage: "The name to greet"})
```

**A** will be the actual function parameter.<br>
**B** is the flag type (notice that it must match the parameter type)
**C** is where the description fields go

The `goat.Describe(T).As(Flag[T])` syntax is used to enforce matching types using generics.
During code generation, those calls are parsed to extract the flag information.
At runtime, both `Describe` and `As` are essentially no-ops.

#### Flag Types

For each argument type (excepting `bool`) we have 3 flag types:

- `RequiredXXXFlag`: Requires that the flag be provided by the user. 
    This is what you get if you don't describe your flags. 
- `DefaultXXXFlag`: Provides a default value, and allows the user to change it.
- `OptionalXXXFlag`: The only option for pointer variables.
    Will ensure that `nil` is received if the flag was not set by the user.

For `bool` arguments, we support only `BoolFlag`. 
It is a `Default` style flag, so you can decide whether you go with default-true or default-false. 

## Dependencies

Goat currently uses [urfave/cli](https://github.com/urfave/cli) for parsing flags.
Other than that, the generated code currently only depends on the standard library.

In the future, I plan to write backends for other popular flag-parsing libraries
(namely [Cobra](https://cobra.dev/) and [flag](https://pkg.go.dev/flag)) so that
users can choose what they depend on.

---

### Internals

See [this blog post](https://blog.tamir.dev/posts/goat-codegen-initial/) for more implementation details.

[Typer]:https://typer.tiangolo.com/