package flags

import "github.com/urfave/cli/v2"

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
