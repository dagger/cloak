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
}
