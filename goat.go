package goat

import (
	"github.com/urfave/cli"
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

func Run(f any) error {
	config := registry[reflect.ValueOf(f)]

	app := &cli.App{
		Flags:  config.Flags,
		Action: config.Action,
	}

	return app.Run(os.Args)
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
