package main

import (
	"fmt"
	"github.com/tmr232/goat"
	"github.com/tmr232/goat/flags"
	"github.com/urfave/cli/v2"
	"go/types"
	"strconv"
	"strings"
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

func (h *hexNumber) Set(value string) error {
	num, err := strconv.ParseUint(strings.TrimPrefix(strings.ToLower(value), "0x"), 16, 32)
	*h = hexNumber(num)
	return err
}

func (h *hexNumber) String() string {
	return fmt.Sprintf("0x%X", *h)
}

func withHex(hex hexNumber) {
	goat.Flag(hex).Default(hexNumber(0x12))
	fmt.Println(hex)
}

func init() {
	flags.RegisterTypeHandler[hexNumber](&customTypeHandler{
		make: func(name, usage string, defaultValue any) cli.Flag {
			if defaultValue == nil {
				val := hexNumber(0)
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
			return *c.Generic(name).(*hexNumber)
		},
	})
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
	//goat.Run(withCustomType)
}
