package python

import (
	"fmt"
	"testing"
)

func TestSetAttr(t *testing.T) {
	obj := struct {
		Name string
		Sub  struct{ Thing int }
	}{"World", struct{ Thing int }{3}}
	fmt.Println(SetAttr(&obj, "Name", "Tamir"))
	fmt.Println(SetAttr(&obj, "Sub.Thing", 7))
	fmt.Println(obj)
}

type Small struct {
	Name  string
	Value int
}
type Thing struct {
	A Small
	B Small
}

type Pattern struct {
	_ Thing
	A struct {
		_     Small
		Value int
	}
	B struct {
		Small
		Name string
	}
}

func TestStructMatch(t *testing.T) {
	fmt.Println(StructMatch[Pattern](Thing{A: Small{"A", 0}, B: Small{"B", 1}}))
}

func TestStructQuery(t *testing.T) {
	query := StructQuery[Pattern](Thing{A: Small{"A", 5}, B: Small{"B", 1}})
	fmt.Println(query.A.Value, query.B.Name, query.B.Value)
}
