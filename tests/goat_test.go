package tests

import (
	"github.com/tmr232/goat"
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
