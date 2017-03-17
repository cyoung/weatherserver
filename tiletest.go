package main

import (
	"fmt"
	"github.com/buckhx/tiles"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("%s <lat> <lng>\n", os.Args[0])
		return
	}

	lat, err := strconv.ParseFloat(os.Args[1], 64)
	if err != nil {
		fmt.Printf("invalid lat: %s\n", err.Error())
		return
	}
	lng, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Printf("invalid lng: %s\n", err.Error())
		return
	}

	myTile := tiles.FromCoordinate(lat, lng, 5)

	url := fmt.Sprintf("https://mesonet.agron.iastate.edu/cache/tile.py/1.0.0/nexrad-n0q-900913/%d/%d/%d.png", myTile.Z, myTile.X, myTile.Y)

	fmt.Printf("%v\n", myTile)
	fmt.Printf("%s\n", url)
}
