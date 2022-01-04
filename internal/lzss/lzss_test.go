package lzss

import (
	"fmt"
	"testing"
)

func TestContains(t *testing.T) {
	buf := []byte("ciao a tutti")
	occ := []byte("tut")
	fmt.Println(contains(buf, occ))
	occ = []byte("ti")
	fmt.Println(contains(buf, occ))
}
