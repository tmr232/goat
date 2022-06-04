package main

import (
	"fmt"
	"github.com/tmr232/goat"
)

//go:generate go run github.com/tmr232/goat/cmd/goat

func app(name string, goodbye bool, question *string) {
	goat.Describe(name).As(goat.DefaultStringFlag{Usage: "The name to greet", Default: "World"})
	goat.Describe(goodbye).As(goat.BoolFlag{Usage: "Enable to say Goodbye", Name: "bye"})
	goat.Describe(question).As(goat.OptionalStringFlag{Usage: "Instead of a greeting, ask a question."})

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

func main() {
	goat.Run(app)
}
