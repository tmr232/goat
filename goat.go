package goat

import (
	"github.com/urfave/cli"
	"log"
	"os"
	"reflect"
)

type RunConfig struct {
	Flags  []cli.Flag
	Action cli.ActionFunc
}

var registry map[reflect.Value]RunConfig

func init() {
	registry = make(map[reflect.Value]RunConfig)
}

func Register(app any, config RunConfig) {
	registry[reflect.ValueOf(app)] = config
}

func RunE(f any) error {
	config := registry[reflect.ValueOf(f)]

	app := &cli.App{
		Flags:  config.Flags,
		Action: config.Action,
	}

	return app.Run(os.Args)
}

func Run(f any) {
	err := RunE(f)
	if err != nil {
		log.Fatal(err)
	}
}

func Command(name string, f any) cli.Command {
	config := registry[reflect.ValueOf(f)]

	return cli.Command{
		Flags:  config.Flags,
		Action: config.Action,
		Name:   name,
	}
}

type Application struct{ *cli.App }

func (app Application) RunE() error {
	return app.App.Run(os.Args)
}
func (app Application) Run() {
	err := app.App.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
func App(name string, commands ...cli.Command) Application {
	return Application{
		App: &cli.App{
			Name:     name,
			Commands: commands,
		},
	}
}

func Flag(any) FluentFlag {
	return FluentFlag{}
}

type FluentFlag struct{}

func (f FluentFlag) Name(string) FluentFlag {
	return FluentFlag{}
}

func (f FluentFlag) Usage(string) FluentFlag {
	return FluentFlag{}
}
func (f FluentFlag) Default(any) FluentFlag {
	return FluentFlag{}
}

type FluentSelf struct{}

func (s FluentSelf) Name(string) FluentSelf {
	return FluentSelf{}
}
func (s FluentSelf) Usage(string) FluentSelf {
	return FluentSelf{}
}

func Self() FluentSelf {
	return FluentSelf{}
}
