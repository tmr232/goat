Template for optional variable generation is below.
This is going to be a bit more involved.

We'll need to keep track of pointer-arguments and ensure we 
generate the variables properly.
After that, we also need to add extra checks for presence. 
But overall, this is still relatively simple.

```go
package main

import (
	"github.com/tmr232/goat"
	"github.com/urfave/cli"
	"log"
	"os"
)

func init() {

	goat.Register(app, func() {

		var goodbye bool
		var name string
		var question string

		__goatApp := &cli.App{
			Flags: []cli.Flag{

				goat.BoolFlag{Usage: "Enable to say Goodbye"}.AsCliFlag("goodbye", &goodbye),
				goat.StringFlag{Usage: "The name to greet"}.AsCliFlag("name", &name),
				goat.OptStringFlag{Usage: "Instead of a greeting, ask a question."}.AsCliFlag("question", &question),
			},
			Action: func(c *cli.Context) {
				var questionArg *string
				if c.IsSet("question") {
					questionArg = &question
				}
				app(name, goodbye, questionArg)
			},
		}

		__err := __goatApp.Run(os.Args)
		if __err != nil {
			log.Fatal(__err)
		}

	})

}

```