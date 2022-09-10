package flags

import (
	"reflect"
	"testing"
)

func TestMakeFlag(t *testing.T) {
	type args struct {
		name         string
		usage        string
		defaultValue any
	}
	tests := []struct {
		name string
		args args
		want Flag
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MakeFlag(tt.args.name, tt.args.usage, tt.args.defaultValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}
