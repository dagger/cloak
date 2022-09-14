package nogen

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func TestSchema(t *testing.T) {
	h := Hugo{}

	s, err := Schema(h)
	if err != nil {
		t.Fatal(err)
	}

	ps, err := Schema(&h)
	if err != nil {
		t.Fatal(err)
	}
	if s != ps {
		t.Fatalf(`generated schema mismatch:
normal:
	%v
pointer:
	%v`, s, ps)
	}

	//t.Log(s)
}

type T struct{}

func (t T) Func1(ctx context.Context, a, b int) (string, error) {
	return "", nil
}

func (t T) MustFunc1(a, b int) string {
	return ""
}

func ExampleSchema() {
	t := T{}

	s, err := Schema(t)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(s)
	// Output:
	// t {
	// 		func1(i: int, i: int): string
	// 		mustfunc1(i: int, i: int): string
	// }
}
