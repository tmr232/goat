package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/tmr232/goat"
)

//go:generate go run github.com/tmr232/goat/cmd/goater

func app(name string, goodbye bool, question *string, times int) error {
	goat.Self().
		Name("application").
		Usage("usage")
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
	return nil
}

func hello() error {
	goat.Self().
		Usage("Say hello to the world.")
	fmt.Println("Hello, World!")
	return nil
}

func greet(name string) error {
	fmt.Printf("Hello, %s!\n", name)
	return nil
}

func fail(msg string) error {
	goat.Flag(msg).Default("default error")
	return errors.New(msg)
}

func main() {
	goat.App("greeter",
		goat.Command(hello),
		goat.Command(greet),
		goat.Command(fail),
		goat.Command(app),
	).Run()
}
