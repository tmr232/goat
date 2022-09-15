package main

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	type args struct {
		slice []int
		op    func(int) string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"nil slice", args{nil, strconv.Itoa}, make([]string, 0)},
		{"empty slice", args{make([]int, 0), strconv.Itoa}, make([]string, 0)},
		{"single item", args{[]int{1}, strconv.Itoa}, []string{"1"}},
		{"multiple items", args{[]int{1, 2, 3}, strconv.Itoa}, []string{"1", "2", "3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Map(tt.args.slice, tt.args.op); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapE(t *testing.T) {
	expectedError := errors.New("My expected error!")
	makeOp := func(errorOn int) func(int) (string, error) {
		return func(i int) (string, error) {
			if i == errorOn {
				return "", expectedError
			}
			return strconv.Itoa(i), nil
		}
	}
	type args struct {
		slice []int
		op    func(int) (string, error)
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{"nil slice", args{nil, makeOp(0)}, make([]string, 0), false},
		{"empty slice", args{make([]int, 0), makeOp(0)}, make([]string, 0), false},
		{"single item, no error", args{[]int{1}, makeOp(0)}, []string{"1"}, false},
		{"single item, error", args{[]int{1}, makeOp(1)}, nil, true},
		{"multiple items, no error", args{[]int{1, 2, 3}, makeOp(0)}, []string{"1", "2", "3"}, false},
		{"multiple items, error", args{[]int{1, 2, 3}, makeOp(2)}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapE(tt.args.slice, tt.args.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapE() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReversed(t *testing.T) {
	type args struct {
		slice []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"nil slice", args{nil}, make([]int, 0)},
		{"empty slice", args{make([]int, 0)}, make([]int, 0)},
		{"single item", args{[]int{1}}, []int{1}},
		{"multiple items", args{[]int{1, 2, 3, 4, 5}}, []int{5, 4, 3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reversed(tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reversed() = %v, want %v", got, tt.want)
			}
		})
	}
}
