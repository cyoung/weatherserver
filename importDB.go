package main

import (
	"./ADDS"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("%s <in.csv> <out.sqlite3>\n", os.Args[0])
		return
	}

	err := ADDS.ImportCSVToNewSQLite(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	} else {
		fmt.Printf("success!\n")
	}
}
