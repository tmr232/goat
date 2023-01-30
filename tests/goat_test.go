package tests

import (
	"bytes"
	"github.com/approvals/go-approval-tests"
	"github.com/tmr232/goat"
	"strings"
	"testing"
)

func Args(args ...string) []string {
	return append([]string{"app"}, args...)
}

func Test_things(t *testing.T) {
	type args struct {
		f    any
		args []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"noFlags", args{noFlags, Args()}, true},
		{"intFlag with valid flag", args{intFlag, Args("--flag", "1")}, true},
		{"intFlag with invalid flag", args{intFlag, Args("--flag", "a")}, false},
		{"intFlag without flag", args{intFlag, Args()}, false},
		{"renamedFlag with correct name", args{renamedFlag, Args("--flag", "1")}, true},
		{"renamedFlag with wrong (original) name", args{renamedFlag, Args("--bla", "1")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := goat.RunWithArgsE(tt.args.f, tt.args.args); tt.want != (got == nil) {
				t.Errorf("goat.Run(...) = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_apps(t *testing.T) {
	type args struct {
		f    any
		args []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"noFlags", args{noFlags, Args()}},
		{"noFlags --help", args{noFlags, Args("--help")}},
		{"intFlag --flag 1", args{intFlag, Args("--flag", "1")}},
		{"intFlag --flag a", args{intFlag, Args("--flag", "a")}},
		{"intFlag", args{intFlag, Args()}},
		{"intFlag --help", args{intFlag, Args("--help")}},
		{"renamedFlag --flag 1", args{renamedFlag, Args("--flag", "1")}},
		{"renamedFlag --bla 1", args{renamedFlag, Args("--bla", "1")}},
		{"Documented --help", args{Documented, Args("--help")}},
		{"flagUsage --help", args{flagUsage, Args("--help")}},
		{"defaultValue --help", args{defaultValue, Args("--help")}},
		{"defaultValue --num 8", args{defaultValue, Args("--num", "8")}},
		{"defaultValue", args{defaultValue, Args()}},
		{"optionalFlag --help", args{optionalFlag, Args("--help")}},
		{"optionalFlag", args{optionalFlag, Args()}},
		{"optionalFlag --num 10", args{optionalFlag, Args("--num", "10")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := goat.FuncToApp(tt.args.f)
			stdout := &bytes.Buffer{}
			defer withWriter(stdout).restore()
			app.Writer = stdout
			_ = app.Run(tt.args.args)
			approvals.Verify(t, stdout)
		})
	}
}

func Test_subcommands(t *testing.T) {
	app := goat.App("test-app", goat.Command(noFlags),
		goat.Command(intFlag),
		goat.Command(renamedFlag),
		goat.Command(Documented),
		goat.Command(flagUsage),
		goat.Command(defaultValue))

	tests := []string{
		"noFlags",
		"noFlags --help",
		"intFlag --flag 1",
		"intFlag --flag a",
		"intFlag",
		"intFlag --help",
		"renamedFlag --flag 1",
		"renamedFlag --bla 1",
		"Documented --help",
		"flagUsage --help",
		"defaultValue --help",
		"defaultValue --num 8",
		"defaultValue",
	}
	for _, tt := range tests {
		args := append([]string{"test-app"}, strings.Split(tt, " ")...)
		stdout := &bytes.Buffer{}
		stdout.WriteString(strings.Join(args, " ") + "\n")
		stdout.WriteString("-----------------------------------------------------\n\n")
		app.Writer = stdout
		t.Run(tt, func(t *testing.T) {
			_ = app.RunWithArgsE(args)
			approvals.Verify(t, stdout)
		})
	}
}

func TestApp(t *testing.T) {
	for _, cmd := range appCmds {
		args := append([]string{"test-app"}, strings.Split(cmd, " ")...)
		stdout := &bytes.Buffer{}
		stdout.WriteString(strings.Join(args, " ") + "\n")
		stdout.WriteString("-----------------------------------------------------\n\n")
		app := getApp(stdout, nil)
		t.Run(cmd, func(t *testing.T) {
			_ = app.RunWithArgsE(args)
			approvals.Verify(t, stdout)
		})
	}
}
