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

type RequiredStringFlag struct {
	Name  string
	Usage string
}

type StringFlag = RequiredStringFlag

func (flag RequiredStringFlag) flag(string) {}
func (flag RequiredStringFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}

func (flag RequiredStringFlag) AsCliFlag(defaultName string, dest *string) cli.Flag {
	return &cli.StringFlag{
		Name:        flag.CliName(defaultName),
		Usage:       flag.Usage,
		Required:    true,
		Destination: dest,
	}
}

type DefaultStringFlag struct {
	Name    string
	Usage   string
	Default string
}

func (flag DefaultStringFlag) flag(string) {}
func (flag DefaultStringFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}

func (flag DefaultStringFlag) AsCliFlag(defaultName string, dest *string) cli.Flag {
	return &cli.StringFlag{
		Name:        flag.CliName(defaultName),
		Usage:       flag.Usage,
		Value:       flag.Default,
		Destination: dest,
	}
}

type OptionalStringFlag struct {
	Name  string
	Usage string
}

func (flag OptionalStringFlag) flag(*string) {}
func (flag OptionalStringFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}

func (flag OptionalStringFlag) AsCliFlag(defaultName string, dest *string) cli.Flag {
	return &cli.StringFlag{
		Name:        flag.CliName(defaultName),
		Usage:       flag.Usage,
		Required:    false,
		Destination: dest,
	}
}

type RequiredIntFlag struct {
	Name  string
	Usage string
}

type IntFlag = RequiredStringFlag

func (flag RequiredIntFlag) flag(int) {}
func (flag RequiredIntFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}
func (flag RequiredIntFlag) AsCliFlag(defaultName string, dest *int) cli.Flag {
	return &cli.IntFlag{
		Name:        flag.CliName(defaultName),
		Usage:       flag.Usage,
		Required:    true,
		Destination: dest,
	}
}

type DefaultIntFlag struct {
	Name    string
	Usage   string
	Default int
}

func (flag DefaultIntFlag) flag(int) {}
func (flag DefaultIntFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}
func (flag DefaultIntFlag) AsCliFlag(defaultName string, dest *int) cli.Flag {
	return &cli.IntFlag{
		Name:        flag.CliName(defaultName),
		Usage:       flag.Usage,
		Value:       flag.Default,
		Destination: dest,
	}
}

type OptionalIntFlag struct {
	Name  string
	Usage string
}

func (flag OptionalIntFlag) flag(*int) {}
func (flag OptionalIntFlag) CliName(defaultName string) string {
	if flag.Name == "" {
		return defaultName
	}
	return flag.Name
}
func (flag OptionalIntFlag) AsCliFlag(defaultName string, dest *int) cli.Flag {
	return &cli.IntFlag{
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
