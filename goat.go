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
	CliName(defaultName string) string
}

type BoolFlag struct {
	Name    string
	Usage   string
	Default bool
}

func (flag BoolFlag) flag(bool) {}
func (flag BoolFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}

func (flag BoolFlag) AsCliFlag(defaultName string, dest *bool) cli.Flag {
	if flag.Default {
		return &cli.BoolTFlag{
			Name:        flag.CliName(defaultName),
			Usage:       flag.Usage,
			Destination: dest,
		}
	} else {
		return &cli.BoolFlag{
			Name:        flag.CliName(defaultName),
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
func (flag StringFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}

func (flag StringFlag) AsCliFlag(defaultName string, dest *string) cli.Flag {
	return &cli.StringFlag{
		Name:        flag.CliName(defaultName),
		Usage:       flag.Usage,
		Value:       flag.Default,
		Required:    true,
		Destination: dest,
	}
}

type OptStringFlag struct {
	Name  string
	Usage string
}

func (flag OptStringFlag) flag(*string) {}
func (flag OptStringFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}

func (flag OptStringFlag) AsCliFlag(defaultName string, dest *string) cli.Flag {
	return &cli.StringFlag{
		Name:        flag.CliName(defaultName),
		Usage:       flag.Usage,
		Required:    false,
		Destination: dest,
	}
}

type Descriptor[T any] struct{}

func (d Descriptor[T]) As(Flag[T]) {}

func Describe[T any](arg T) Descriptor[T] { return Descriptor[T]{} }

func GetOptional[T any](c *cli.Context, name string, dest *T) *T {
	if c.IsSet(name) {
		return dest
	}
	return nil
}
