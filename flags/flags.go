package flags

import (
	cli "github.com/urfave/cli/v2"
)

func tryDeref[T any](ptr *T) T {
	if ptr == nil {
		return *new(T)
	}
	return *ptr
}

type Flag interface {
	AsCliFlag() cli.Flag
}

type DefaultIntFlag struct {
	Name    string
	Usage   string
	Default int
}

func (flag DefaultIntFlag) AsCliFlag() cli.Flag {
	return &cli.IntFlag{
		Name:  flag.Name,
		Usage: flag.Usage,
		Value: flag.Default,
	}
}

type RequiredIntFlag struct {
	Name  string
	Usage string
}

func (flag RequiredIntFlag) AsCliFlag() cli.Flag {
	return &cli.IntFlag{
		Name:     flag.Name,
		Usage:    flag.Usage,
		Required: true,
	}
}

type OptionalIntFlag struct {
	Name  string
	Usage string
}

func (flag OptionalIntFlag) AsCliFlag() cli.Flag {
	return &cli.IntFlag{
		Name:     flag.Name,
		Usage:    flag.Usage,
		Required: false,
	}
}

type DefaultStringFlag struct {
	Name    string
	Usage   string
	Default string
}

func (flag DefaultStringFlag) AsCliFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  flag.Name,
		Usage: flag.Usage,
		Value: flag.Default,
	}
}

type RequiredStringFlag struct {
	Name  string
	Usage string
}

func (flag RequiredStringFlag) AsCliFlag() cli.Flag {
	return &cli.StringFlag{
		Name:     flag.Name,
		Usage:    flag.Usage,
		Required: true,
	}
}

type OptionalStringFlag struct {
	Name  string
	Usage string
}

func (flag OptionalStringFlag) AsCliFlag() cli.Flag {
	return &cli.StringFlag{
		Name:     flag.Name,
		Usage:    flag.Usage,
		Required: false,
	}
}

type BoolFlag struct {
	Name    string
	Usage   string
	Default bool
}

func (flag BoolFlag) AsCliFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:  flag.Name,
		Usage: flag.Usage,
		Value: flag.Default,
	}
}

func tryCast[T any](from any) T {
	if from == nil {
		return *new(T)
	}
	return from.(T)
}

func MakeFlag[T any](name string, usage string, defaultValue any) Flag {
	switch any(*new(T)).(type) {
	case int:
		if defaultValue == nil {
			return RequiredIntFlag{Name: name, Usage: usage}
		}
		return DefaultIntFlag{
			Name:    name,
			Usage:   usage,
			Default: tryCast[int](defaultValue),
		}
	case *int:
		return OptionalIntFlag{
			Name:  name,
			Usage: usage,
		}
	case string:
		if defaultValue == nil {
			return RequiredStringFlag{
				Name:  name,
				Usage: usage,
			}
		}
		return DefaultStringFlag{
			Name:    name,
			Usage:   usage,
			Default: tryCast[string](defaultValue),
		}
	case *string:
		return OptionalStringFlag{
			Name:  name,
			Usage: usage,
		}
	case bool:
		return BoolFlag{
			Name:    name,
			Usage:   usage,
			Default: tryCast[bool](defaultValue),
		}
	default:
		panic("Missing handler for type")
	}

}

func cast[T any](from any) T {
	return from.(T)
}

func GetFlag[T any](c *cli.Context, name string) T {
	switch any(*new(T)).(type) {
	case int:
		return cast[T](c.Int(name))
	case *int:
		if c.IsSet(name) {
			i := c.Int(name)
			return cast[T](&i)
		}
		return *new(T)
	case string:
		return cast[T](c.String(name))
	case *string:
		if c.IsSet(name) {
			s := c.String(name)
			return cast[T](&s)
		}
		return *new(T)
	case bool:
		return cast[T](c.Bool(name))
	}
	panic("oh no!")
}

/*
The codegen part will create the calls to `MakeXXXFlag` with the correct `FlagDescription` struct to pass in.
This is done to allow separating the descriptor fields from the actual struct fields.
If the default is empty, it should be set to `nil` (which will happen automatically :) )

The rest of the codegen (for now) will be as it is today. But hopefully, the structs will be simpler.
Once I have the primitive types in place (numbers, strings, booleans) in place, it'll be time to think about grouping.

Note that avoiding a `Destination` field in the CLI flags will help with generation of command-groups.
It'll mean that every command's generation can be localized, and there's no need for global variables.
*/
