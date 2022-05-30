package main

import (
	"fmt"
	"github.com/tmr232/goat"
)

//go:generate go run github.com/tmr232/goat/cmd/goater

func app(name string, goodbye bool, question *string, times int) {
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
	goat.Run(app)
}
