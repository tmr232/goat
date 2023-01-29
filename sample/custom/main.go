package main

import (
	"github.com/tmr232/goat"
	"github.com/tmr232/goat/flags"
	"github.com/urfave/cli/v2"
	"go/types"
)

type customType int

func withCustomType(num customType, x int, t types.Type) {}

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

func init() {
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
			return c.Int(name)
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

func f() {
	goat.Command(withCustomType)
}

func main() {
	goat.Run(withCustomType)
}
