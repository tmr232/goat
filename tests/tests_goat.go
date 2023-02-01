package tests

import (
	goat "github.com/tmr232/goat"
	flags "github.com/tmr232/goat/flags"
	cli "github.com/urfave/cli/v2"
)

func init() {
	goat.Register(NoFlags, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{},
			Name:  "NoFlags",
			Usage: "has no flags.",
			Action: func(c *cli.Context) error {
				NoFlags()
				return nil
			},
		}
	})

	goat.Register(FlagsWithUsage, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{
				flags.MakeFlag[int]("a", "This is a", nil),
				flags.MakeFlag[int]("b", "Nice!", nil),
				flags.MakeFlag[int]("c", "C.", nil),
			},
			Name:  "FlagsWithUsage",
			Usage: "has usage for its flags!",
			Action: func(c *cli.Context) error {
				FlagsWithUsage(
					flags.GetFlag[int](c, "a"),
					flags.GetFlag[int](c, "b"),
					flags.GetFlag[int](c, "c"),
				)
				return nil
			},
		}
	})

	goat.Register(noFlags, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{},
			Name:  "noFlags",
			Usage: "",
			Action: func(c *cli.Context) error {
				noFlags()
				return nil
			},
		}
	})

	goat.Register(intFlag, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{
				flags.MakeFlag[int]("flag", "", nil),
			},
			Name:  "intFlag",
			Usage: "",
			Action: func(c *cli.Context) error {
				intFlag(
					flags.GetFlag[int](c, "flag"),
				)
				return nil
			},
		}
	})

	goat.Register(renamedFlag, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{
				flags.MakeFlag[int]("flag", "", nil),
			},
			Name:  "renamedFlag",
			Usage: "",
			Action: func(c *cli.Context) error {
				renamedFlag(
					flags.GetFlag[int](c, "flag"),
				)
				return nil
			},
		}
	})

	goat.Register(Documented, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{},
			Name:  "Documented",
			Usage: "has some neat docs!\n\nIt's just so nice to document your code.",
			Action: func(c *cli.Context) error {
				Documented()
				return nil
			},
		}
	})

	goat.Register(flagUsage, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{
				flags.MakeFlag[int]("num", "A number of things.", nil),
				flags.MakeFlag[string]("str", "A piece of text.", nil),
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
		}
	})

	goat.Register(defaultValue, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{
				flags.MakeFlag[int]("num", "", 5),
			},
			Name:  "defaultValue",
			Usage: "",
			Action: func(c *cli.Context) error {
				defaultValue(
					flags.GetFlag[int](c, "num"),
				)
				return nil
			},
		}
	})

	goat.Register(optionalFlag, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{
				flags.MakeFlag[*int]("num", "This flag is optional!", nil),
			},
			Name:  "optionalFlag",
			Usage: "",
			Action: func(c *cli.Context) error {
				optionalFlag(
					flags.GetFlag[*int](c, "num"),
				)
				return nil
			},
		}
	})

	goat.Register(withIntFlags, func() goat.RunConfig {
		return goat.RunConfig{
			Flags: []cli.Flag{
				flags.MakeFlag[int]("required", "", nil),
				flags.MakeFlag[int]("defaultValue", "", 42),
				flags.MakeFlag[*int]("optional", "", nil),
			},
			Name:  "withIntFlags",
			Usage: "",
			Action: func(c *cli.Context) error {
				withIntFlags(
					flags.GetFlag[int](c, "required"),
					flags.GetFlag[int](c, "defaultValue"),
					flags.GetFlag[*int](c, "optional"),
				)
				return nil
			},
		}
	})
}
