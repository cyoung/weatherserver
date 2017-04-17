package main

import (
	"../ADDS"
	"encoding/json"
	"fmt"
	"github.com/kellydunn/golang-geo"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	StationLat          float64
	StationLng          float64
	StationServiceRange float64 // Statute miles.
}

var myConfig Config

var metars map[string]ADDS.ADDSMETAR // station -> ADDSMETAR.
var tafs map[string]ADDS.ADDSTAF     // station -> ADDSTAF.
var weatherMutex *sync.Mutex

var selfGeo *geo.Point

func weatherUpdater() {
	updateTicker := time.NewTicker(10 * time.Minute)
	for {
		weatherMutex.Lock()
		// Update the weather.
		// Get METARs.
		addsMetars, err := ADDS.GetLatestADDSMETARsInRadiusOf(SERVICE_RANGE, selfGeo)
		if err != nil {
			fmt.Printf("error obtaining METARs: %s\n", err.Error())
		} else {
			for _, metar := range addsMetars {
				metars[metar.StationID] = metar
			}
		}
		// Get TAFs.
		addsTafs, err := ADDS.GetLatestADDSTAFsInRadiusOf(SERVICE_RANGE, selfGeo)
		if err != nil {
			fmt.Printf("error obtaining TAFs: %s\n", err.Error())
		} else {
			for _, taf := range addsTafs {
				tafs[taf.StationID] = taf
			}
		}
		weatherMutex.Unlock()
		<-updateTicker.C
	}
}

func main() {
	weatherMutex = &sync.Mutex{}
	metars = make(map[string]ADDS.ADDSMETAR, 0)
	tafs = make(map[string]ADDS.ADDSTAF, 0)

	// Read and parse config file.
	fp, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("Can't open 'config.json'.\n")
		return
	}
	decoder := json.NewDecoder(fp)
	err = decoder.Decode(&myConfig)
	if err != nil {
		fmt.Printf("Couldn't read 'config.json'.\n")
		return
	}

	selfGeo = geo.NewPoint(myConfig.StationLat, myConfig.StationLng)

	go weatherUpdater()

	// Weather broadcast loop.
	broadcastTicker := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-broadcastTicker.C:
			weatherMutex.Lock()
			fmt.Printf("**************************************************\n")
			fmt.Printf("METAR: %d, TAF: %d, TOTAL: %d\n", len(metars), len(tafs), len(metars)+len(tafs))
			for _, metar := range metars {
				t := metar.Text
				if !strings.HasPrefix(t, "METAR ") {
					t = "METAR " + t
				}
				fmt.Printf("%s\n", t)
			}
			for _, taf := range tafs {
				t := taf.Text
				if !strings.HasPrefix(t, "METAR ") {
					t = "TAF " + t
				}
				fmt.Printf("%s\n", t)
			}
			fmt.Printf("**************************************************\n")
			weatherMutex.Unlock()
		}
	}
}
