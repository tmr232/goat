package main

import (
	"fmt"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"github.com/tmr232/goat"
	"github.com/tmr232/goat/flags"
	"github.com/urfave/cli/v2"
)

//go:generate go run github.com/tmr232/goat/cmd/goater
type customType int

func withCustomType(num customType, x int, t types.Type) {
	fmt.Println(num)
}

type customTypeHandler struct {
	make func(name, usage string, defaultValue any) cli.Flag
	get  func(c *cli.Context, name string) any
}

func (h *customTypeHandler) MakeFlag(name, usage string, defaultValue any) cli.Flag {
	return h.make(name, usage, defaultValue)
}

func (h *customTypeHandler) GetFlag(c *cli.Context, name string) any {
	return h.get(c, name)
}

type hexNumber uint64
type mockHex uint64

func (h *hexNumber) Set(value string) error {
	num, err := strconv.ParseUint(strings.TrimPrefix(strings.ToLower(value), "0x"), 16, 32)
	*h = hexNumber(num)
	return err
}

func (h *hexNumber) String() string {
	return fmt.Sprintf("0x%X", *h)
}

func (h *hexNumber) FromString(value string) (any, error) {
	num, err := strconv.ParseUint(strings.TrimPrefix(strings.ToLower(value), "0x"), 16, 32)
	return hexNumber(num), err
}

func withHex(hex mockHex, phex *hexNumber) {
	goat.Flag(hex).Default(hexNumber(0x12))
	fmt.Println(hex, phex)
}

type Custom interface {
	String() string
	FromString(value string) (any, error)
}

type CustomGeneric[T Custom] struct {
	value T
}

func (c *CustomGeneric[T]) Set(value string) error {
	v, err := c.value.FromString(value)
	c.value = v.(T)
	return err
}

func (c *CustomGeneric[T]) String() string {
	return c.value.String()
}

type MyGeneric[T any] struct {
	val        T
	toString   func(*T) string
	fromString func(*T, string) error
}

func (m *MyGeneric[T]) Set(value string) error {
	return m.fromString(&m.val, value)
}

func (m *MyGeneric[T]) String() string {
	return m.toString(&m.val)
}

func registerGenAs[T any, As any]() {
	genericType := reflect.TypeOf(*new(As))

	flags.RegisterTypeHandler[T](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			genericValue := reflect.New(genericType)
			genericInterface := genericValue

			if defaultValue == nil {
				return &cli.GenericFlag{
					Name:     name,
					Usage:    usage,
					Value:    genericInterface.Interface().(cli.Generic),
					Required: true,
				}
			}

			genericInterface = reflect.New(genericType)
			genericInterface.Elem().Set(reflect.ValueOf(defaultValue).Convert(genericType))

			return &cli.GenericFlag{
				Name:  name,
				Usage: usage,
				Value: genericInterface.Interface().(cli.Generic),
			}
		},
		get: func(c *cli.Context, name string) any {
			return reflect.ValueOf(c.Generic(name)).Elem().Convert(reflect.TypeOf(*new(T))).Interface()
		},
	})
}

func registerGen[T any]() {
	genericType := reflect.TypeOf(*new(T))

	flags.RegisterTypeHandler[T](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			genericValue := reflect.New(genericType)
			genericInterface := genericValue

			if defaultValue == nil {
				return &cli.GenericFlag{
					Name:     name,
					Usage:    usage,
					Value:    genericInterface.Interface().(cli.Generic),
					Required: true,
				}
			}

			genericInterface = reflect.New(genericType)
			genericInterface.Elem().Set(reflect.ValueOf(defaultValue))

			return &cli.GenericFlag{
				Name:  name,
				Usage: usage,
				Value: genericInterface.Interface().(cli.Generic),
			}
		},
		get: func(c *cli.Context, name string) any {
			return reflect.ValueOf(c.Generic(name)).Elem().Interface()
		},
	})
}

func registerGenOpt[T any]() {
	genericType := reflect.TypeOf(*new(T))

	flags.RegisterTypeHandler[*T](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			genericValue := reflect.New(genericType)
			genericInterface := genericValue

			if defaultValue == nil {
				return &cli.GenericFlag{
					Name:  name,
					Usage: usage,
					Value: genericInterface.Interface().(cli.Generic),
				}
			}

			genericInterface = reflect.New(genericType)
			genericInterface.Elem().Set(reflect.ValueOf(defaultValue))

			return &cli.GenericFlag{
				Name:  name,
				Usage: usage,
				Value: genericInterface.Interface().(cli.Generic),
			}
		},
		get: func(c *cli.Context, name string) any {
			if c.IsSet(name) {
				return reflect.ValueOf(c.Generic(name)).Interface()
			}
			return nil
		},
	})
}

func registerGeneric[T any](toString func(*T) string, fromString func(*T, string) error) {
	flags.RegisterTypeHandler[T](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			val := MyGeneric[T]{
				toString:   toString,
				fromString: fromString,
			}
			if defaultValue == nil {
				return &cli.GenericFlag{
					Name:     name,
					Usage:    usage,
					Value:    &val,
					Required: true,
				}
			}
			val.val = defaultValue.(T)
			return &cli.GenericFlag{
				Name:  name,
				Usage: usage,
				Value: &val,
			}
		},
		get: func(c *cli.Context, name string) any {
			return c.Generic(name).(*MyGeneric[T]).val
		},
	})
}

//func registerGeneric[T cli.Generic]() {
//	genericType := reflect.TypeOf(*new(T))
//	if genericType.Kind() == reflect.Pointer {
//		value := reflect.
//	}
//}

func registerCustom[As any, T Custom]() {
	flags.RegisterTypeHandler[As](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			if defaultValue == nil {
				var val CustomGeneric[T]
				return &cli.GenericFlag{
					Name:     name,
					Usage:    usage,
					Value:    &val,
					Required: true,
				}
			}
			val := defaultValue.(hexNumber)
			return &cli.GenericFlag{
				Name:  name,
				Usage: usage,
				Value: &val,
			}
		},
		get: func(c *cli.Context, name string) any {
			return *c.Generic(name).(*As)
		},
	})
}

func init() {
	// registerCustom[hexNumber, *hexNumber]()
	registerGen[hexNumber]()
	registerGenAs[mockHex, hexNumber]()
	registerGenOpt[hexNumber]()
	//registerGeneric[hexNumber](
	//	func(t *hexNumber) string {
	//		return fmt.Sprintf("0x%X", *t)
	//	},
	//	func(h *hexNumber, s string) error {
	//		num, err := strconv.ParseUint(strings.TrimPrefix(strings.ToLower(s), "0x"), 16, 32)
	//		*h = hexNumber(num)
	//		return err
	//	},
	//)
	//flags.RegisterTypeHandler[hexNumber](&customTypeHandler{
	//	make: func(name, usage string, defaultValue any) cli.Flag {
	//		if defaultValue == nil {
	//			val := hexNumber(0)
	//			return &cli.GenericFlag{
	//				Name:     name,
	//				Usage:    usage,
	//				Value:    &val,
	//				Required: true,
	//			}
	//		}
	//		val := defaultValue.(hexNumber)
	//		return &cli.GenericFlag{
	//			Name:  name,
	//			Usage: usage,
	//			Value: &val,
	//		}
	//	},
	//	get: func(c *cli.Context, name string) any {
	//		return *c.Generic(name).(*hexNumber)
	//	},
	//})
	flags.RegisterTypeHandler[customType](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			if defaultValue == nil {
				return &cli.IntFlag{
					Name:     name,
					Usage:    usage,
					Required: true,
				}
			}
			return &cli.IntFlag{
				Name:  name,
				Usage: usage,
				Value: 5,
			}
		},
		get: func(c *cli.Context, name string) any {
			return customType(c.Int(name))
		},
	})
	flags.RegisterTypeHandler[types.Type](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			if defaultValue == nil {
				return &cli.IntFlag{
					Name:     name,
					Usage:    usage,
					Required: true,
				}
			}
			return &cli.IntFlag{
				Name:  name,
				Usage: usage,
				Value: 5,
			}
		},
		get: func(c *cli.Context, name string) any {
			return nil
		},
	})
}

func main() {
	goat.Run(withHex)
	// goat.Run(withCustomType)
}
