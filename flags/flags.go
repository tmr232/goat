package flags

import (
	cli "github.com/urfave/cli/v2"
	"reflect"
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

type SimpleFlag struct {
	Flag cli.Flag
}

func (f *SimpleFlag) AsCliFlag() cli.Flag {
	return f.Flag
}

func AsSimpleFlag(flag cli.Flag) Flag {
	return &SimpleFlag{flag}
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

type TypeHandler interface {
	MakeFlag(name, usage string, defaultValue any) Flag
	GetFlag(c *cli.Context, name string) any
}

var flagHandlers map[reflect.Type]TypeHandler

func init() {
	flagHandlers = make(map[reflect.Type]TypeHandler)
}

func RegisterTypeHandler[T any](handler TypeHandler) {
	handledType := reflect.TypeOf(*new(T))
	_, exists := flagHandlers[handledType]
	if exists {
		panic("Type handler for type " + handledType.Name() + " already exists.")
	}
	flagHandlers[handledType] = handler
}

type typeHandlerImpl struct {
	makeFlag func(name, usage string, defaultValue any) Flag
	getFlag  func(c *cli.Context, name string) any
}

func (impl *typeHandlerImpl) MakeFlag(name, usage string, defaultValue any) Flag {
	return impl.makeFlag(name, usage, defaultValue)
}

func (impl *typeHandlerImpl) GetFlag(c *cli.Context, name string) any {
	return impl.getFlag(c, name)
}

func init() {
	RegisterTypeHandler[int](&typeHandlerImpl{
		makeFlag: func(name, usage string, defaultValue any) Flag {
			if defaultValue == nil {
				return RequiredIntFlag{Name: name, Usage: usage}
			}
			return DefaultIntFlag{
				Name:    name,
				Usage:   usage,
				Default: tryCast[int](defaultValue),
			}
		},
		getFlag: func(c *cli.Context, name string) any {
			return c.Int(name)
		},
	})

	RegisterTypeHandler[*int](&typeHandlerImpl{
		makeFlag: func(name, usage string, defaultValue any) Flag {
			return OptionalIntFlag{
				Name:  name,
				Usage: usage,
			}
		},
		getFlag: func(c *cli.Context, name string) any {
			if c.IsSet(name) {
				i := c.Int(name)
				return &i
			}
			return nil
		},
	})
	RegisterTypeHandler[string](&typeHandlerImpl{
		makeFlag: func(name, usage string, defaultValue any) Flag {
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
		},
		getFlag: func(c *cli.Context, name string) any {
			return c.String(name)
		},
	})
	RegisterTypeHandler[*string](&typeHandlerImpl{
		makeFlag: func(name, usage string, defaultValue any) Flag {
			return OptionalStringFlag{
				Name:  name,
				Usage: usage,
			}
		},
		getFlag: func(c *cli.Context, name string) any {
			if c.IsSet(name) {
				s := c.String(name)
				return &s
			}
			return nil
		},
	})
	RegisterTypeHandler[bool](&typeHandlerImpl{
		makeFlag: func(name, usage string, defaultValue any) Flag {
			return BoolFlag{
				Name:    name,
				Usage:   usage,
				Default: tryCast[bool](defaultValue),
			}
		},
		getFlag: func(c *cli.Context, name string) any {
			return c.Bool(name)
		},
	})
}

// MakeFlag creates a flag from a type and description values.
func MakeFlag[T any](name string, usage string, defaultValue any) Flag {
	handler, exists := flagHandlers[reflect.TypeOf(*new(T))]
	if !exists {
		panic("Missing handler for type")
	}
	return handler.MakeFlag(name, usage, defaultValue)
}

func GetFlag[T any](c *cli.Context, name string) T {
	handler, exists := flagHandlers[reflect.TypeOf(*new(T))]
	if !exists {
		panic("oh no!")
	}
	return handler.GetFlag(c, name).(T)
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
