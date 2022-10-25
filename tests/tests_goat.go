package tests

import (
	"github.com/tmr232/goat"
	"github.com/tmr232/goat/flags"
	"github.com/urfave/cli/v2"
)

func init() {
	goat.Register(noFlags, goat.RunConfig{
		Flags: []cli.Flag{},
		Name:  "noFlags",
		Usage: "",
		Action: func(c *cli.Context) error {
			noFlags()
			return nil
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			return cflags
		},
	})

	goat.Register(intFlag, goat.RunConfig{
		Flags: []cli.Flag{
			flags.MakeFlag[int]("flag", "", nil).AsCliFlag(),
		},
		Name:  "intFlag",
		Usage: "",
		Action: func(c *cli.Context) error {
			intFlag(
				flags.GetFlag[int](c, "flag"),
			)
			return nil
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			cflags["flag"] = flags.GetFlag[int](c, "flag")
			return cflags
		},
	})

	goat.Register(renamedFlag, goat.RunConfig{
		Flags: []cli.Flag{
			flags.MakeFlag[int]("flag", "", nil).AsCliFlag(),
		},
		Name:  "renamedFlag",
		Usage: "",
		Action: func(c *cli.Context) error {
			renamedFlag(
				flags.GetFlag[int](c, "flag"),
			)
			return nil
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			cflags["flag"] = flags.GetFlag[int](c, "flag")
			return cflags
		},
	})

	goat.Register(Documented, goat.RunConfig{
		Flags: []cli.Flag{},
		Name:  "Documented",
		Usage: "has some neat docs!\n\nIt's just so nice to document your code.",
		Action: func(c *cli.Context) error {
			Documented()
			return nil
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			return cflags
		},
	})

	goat.Register(flagUsage, goat.RunConfig{
		Flags: []cli.Flag{
			flags.MakeFlag[int]("num", "A number of things.", nil).AsCliFlag(),
			flags.MakeFlag[string]("str", "A piece of text.", nil).AsCliFlag(),
		},
		Name:  "flagUsage",
		Usage: "",
		Action: func(c *cli.Context) error {
			flagUsage(
				flags.GetFlag[int](c, "num"),
				flags.GetFlag[string](c, "str"),
			)
			return nil
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			cflags["num"] = flags.GetFlag[int](c, "num")
			cflags["str"] = flags.GetFlag[string](c, "str")
			return cflags
		},
	})

	goat.Register(defaultValue, goat.RunConfig{
		Flags: []cli.Flag{
			flags.MakeFlag[int]("num", "", 5).AsCliFlag(),
		},
		Name:  "defaultValue",
		Usage: "",
		Action: func(c *cli.Context) error {
			defaultValue(
				flags.GetFlag[int](c, "num"),
			)
			return nil
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			cflags["num"] = flags.GetFlag[int](c, "num")
			return cflags
		},
	})
}
