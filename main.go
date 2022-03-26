package main

import (
	"fmt"
	"github.com/tmr232/goat/goat"
	"log"
	"os"
)

type BaseArgs struct {
	Flag *bool
}

type HelloArgs struct {
	BaseArgs
	Name *string `name:"name"`
}

type HelloError string

func (h HelloError) Error() string {
	return string(h)
}

func Hello(args HelloArgs) error {
	fmt.Println("Hello ", args.Name, args.Flag)
	return HelloError("Error!!!")
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
		goat.Usage("Greets things & people."),
		goat.Command("hello", Hello),
		goat.Command("goodbye", Goodbye, goat.Usage("Says goodbye.")),
		goat.Group("say",
			goat.Usage("Says many things."),
			goat.Command("hello", Hello),
			goat.Command("goodbye", Goodbye),
		),
	)

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
