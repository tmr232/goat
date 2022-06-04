package goat

import (
	"github.com/urfave/cli"
	"reflect"
)

var registry map[reflect.Value]func()

func init() {
	registry = make(map[reflect.Value]func())
}

func Register(app any, wrapper func()) {
	registry[reflect.ValueOf(app)] = wrapper
}

func Run(f any) {
	registry[reflect.ValueOf(f)]()
}

type Flag[T any] interface {
	flag(T)
}

type BoolFlag struct {
	Name    string
	Usage   string
	Default bool
}

func (flag BoolFlag) flag(bool) {}

func (flag BoolFlag) AsCliFlag(defaultName string, dest *bool) cli.Flag {
	name := flag.Name
	if name == "" {
		name = defaultName
	}
	if flag.Default {
		return &cli.BoolTFlag{
			Name:        name,
			Usage:       flag.Usage,
			Destination: dest,
		}
	} else {
		return &cli.BoolFlag{
			Name:        name,
			Usage:       flag.Usage,
			Destination: dest,
		}
	}
}

type StringFlag struct {
	Name    string
	Usage   string
	Default string
}

func (flag StringFlag) flag(string) {}

func (flag StringFlag) AsCliFlag(defaultName string, dest *string) cli.Flag {
	name := flag.Name
	if name == "" {
		name = defaultName
	}
	return &cli.StringFlag{
		Name:        name,
		Usage:       flag.Usage,
		Value:       flag.Default,
		Destination: dest,
	}
}

type Descriptor[T any] struct{}

func (d Descriptor[T]) As(Flag[T]) {}

func Describe[T any](arg T) Descriptor[T] { return Descriptor[T]{} }
