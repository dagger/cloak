package nogen

import "testing"

func TestSchema(t *testing.T) {
	h := Hugo{}

	s, err := Schema(h)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(s)
}
