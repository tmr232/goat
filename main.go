package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tmr232/goat/goat"
)

type BaseArgs struct {
	Flag goat.Optional[bool]
}

type HelloArgs struct {
	BaseArgs `embed:""`
	Name     goat.Optional[string] `name:"name"`
}

func Hello(args HelloArgs) error {
	fmt.Println("Hello ", args.Name)
	return nil
}

type GoodbyeArgs struct {
	Name string `name:"name" usage:"the name to say goodbye to"`
}

func Goodbye(args GoodbyeArgs) error {
	fmt.Println("Goodbye ", args.Name)
	return nil
}

func main() {
	app := goat.App(
		"hello-world",
		goat.Command("hello", Hello),
		goat.Command("goodbye", Goodbye, goat.Usage("Says goodbye.")),
		goat.Group("say",
			goat.Command("hello", Hello),
			goat.Command("goodbye", Goodbye),
			goat.Usage("Says many things."),
		),
		goat.Usage("Greets things & people."),
	)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
