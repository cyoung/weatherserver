package main

import (
	"./ADDS"
	"fmt"
)

func main() {
	db, err := ADDS.NewAirportDB("./airports.sqlite3")
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		return
	}

	numAirportsInResponse := 0
	sortedAirports := db.FindClosestAirports(42.4768420, -83.6610550)
	for _, a := range sortedAirports {
		//		fmt.Printf("%s, %f\n", a.ThisAirport.Ident, a.Distance)
		if numAirportsInResponse > 50 {
			break
		}
		metar, err := ADDS.GetLatestADDSMETAR(a.ThisAirport.Ident)
		if err != nil {
			fmt.Printf("err: %s\n", err.Error())
			continue // Couldn't get METAR.
		}
		fmt.Printf("%s, %f, %s\n", a.ThisAirport.Ident, a.Distance, metar.FlightCategory)
		numAirportsInResponse++
	}
}
