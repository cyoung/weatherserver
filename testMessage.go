package main

import (
	"./RockBLOCK"
	"fmt"
)

func main() {
	r, err := RockBLOCK.NewRockBLOCKSerial()
	if err != nil {
		fmt.Printf("init error: %s\n", err.Error())
		return
	} else {
		fmt.Printf("initialized\n")
	}

	err = r.SendBinary([]byte("hello"))
	if err != nil {
		fmt.Printf("send error: %s\n", err.Error())
		return
	} else {
		fmt.Printf("sent\n")
	}
}
