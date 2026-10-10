package idgen

import (
	"strings"
	"testing"
)

func TestNewUnique(t *testing.T) {
	a := New("res")
	b := New("res")
	if a == b {
		t.Fatal("expected unique ids")
	}
	if !strings.HasPrefix(a, "res_") {
		t.Fatalf("prefix: %s", a)
	}
}
