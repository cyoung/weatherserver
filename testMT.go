package main

import (
	"./RockBLOCK"
	"fmt"
)

func main() {
	m := new(RockBLOCK.RockBLOCKCOREOutgoing)
	m.IMEI = RockBLOCK.TEST_IMEI
	m.Data = []byte("HELLO!")
	a, err := m.Send()
	if err != nil {
		fmt.Printf("a=%s, err=%s\n", a, err.Error())
	} else {
		fmt.Printf("a=%s\n", a)
	}
}
