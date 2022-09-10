package main

import (
	"fmt"
	"github.com/tmr232/goat"
)

//go:generate go run github.com/tmr232/goat/cmd/goater

func app(name string, goodbye bool, question *string, times int) error {
	goat.Flag(name).
		Usage("The name to greet")
	goat.Flag(goodbye).
		Name("bye").
		Usage("say goodbye")
	goat.Flag(question).
		Usage("instead of greeting, ask a question")
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
	fmt.Println("Hello, World!")
	return nil
}

func greet(name string) error {
	fmt.Printf("Hello, %s!\n", name)
	return nil
}

func main() {
	goat.App("greeter",
		goat.Command("hello", hello),
		goat.Command("greet", greet),
	).Run()
}
