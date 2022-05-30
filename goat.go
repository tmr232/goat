package goat

import (
	"reflect"
)

var registry map[reflect.Value]func()

func init() {
	registry = make(map[reflect.Value]func())
}

func Register(app any, wrapper func()) {
	registry[reflect.ValueOf(app)] = wrapper
}

func Run(f any) {
	registry[reflect.ValueOf(f)]()
}
