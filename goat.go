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
	Name     string
	Usage    string
	Required bool
	Value    bool
}

func (flag BoolFlag) flag(bool) {}

func (flag BoolFlag) AsCliFlag(dest *bool) cli.Flag {
	if flag.Value {
		return &cli.BoolTFlag{
			Name:        flag.Name,
			Usage:       flag.Name,
			Required:    flag.Required,
			Destination: dest,
		}
	} else {
		return &cli.BoolFlag{
			Name:        flag.Name,
			Usage:       flag.Name,
			Required:    flag.Required,
			Destination: dest,
		}
	}
}

type StringFlag struct {
	Name     string
	Usage    string
	Required bool
	Value    string
}

func (flag StringFlag) flag(string) {}

func (flag StringFlag) AsCliFlag(dest *string) cli.Flag {
	return &cli.StringFlag{
		Name:        flag.Name,
		Usage:       flag.Usage,
		Required:    flag.Required,
		Value:       flag.Value,
		Destination: dest,
	}
}

func Describe[T any](arg T, flagDef Flag[T]) {}
