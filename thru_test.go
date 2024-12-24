package thru

import (
	"fmt"
	"testing"
)

func TestWarp(t *testing.T) {
	pt, _ := ParseE("2024-08-31")
	pt2, _ := ParseE("2024-07-30")
	// pt, _ := time.Parse(time.DateOnly, "2023-03-01")
	fmt.Println(pt.DiffIn(pt2, "M"))
}

func TestStart(t *testing.T) {
	pt, _ := ParseE("2024-03-02 15:04:05.999999999")
	fmt.Println(pt.Start())
}
