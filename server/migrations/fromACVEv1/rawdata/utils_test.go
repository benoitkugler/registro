package rawdata

import (
	"fmt"
	"testing"
)

func TestColors(t *testing.T) {
	c := RGBA{200, 5, 10, 150}
	fmt.Println(c.AHex())
	fmt.Println(c.Hex())
	c2 := HexColor("#ff4512")
	fmt.Println(c2.AHex())
	fmt.Println(c2.Hex())
	fmt.Println(NonCommencee.Color().Hex())
}
