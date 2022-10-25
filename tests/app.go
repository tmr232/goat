package tests

import (
	"github.com/tmr232/goat"
	"io"
)

// NoFlags has no flags.
func NoFlags() {}

// FlagsWithUsage has usage for its flags!
func FlagsWithUsage(a, b, c int) {
	goat.Flag(a).Usage("This is a")
	goat.Flag(b).Usage("Nice!")
	goat.Flag(c).Usage("C.")
}

func getApp(stdout, stderr io.Writer) goat.Application {
	app := goat.App("test-app", goat.Command(NoFlags), goat.Command(FlagsWithUsage))
	app.Writer = stdout
	app.ErrWriter = stderr

	return app
}

var appCmds = []string{
	"NoFlags --help",
	"NoFlags --a-flag",
	"FlagsWithUsage --help",
}
