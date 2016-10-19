package main

import (
	"./ADDS"
	"./RockBLOCK"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// From stratux.
type SituationData struct {
	// From GPS.
	LastFixSinceMidnightUTC  float32
	Lat                      float32
	Lng                      float32
	Quality                  uint8
	HeightAboveEllipsoid     float32 // GPS height above WGS84 ellipsoid, ft. This is specified by the GDL90 protocol, but most EFBs use MSL altitude instead. HAE is about 70-100 ft below GPS MSL altitude over most of the US.
	GeoidSep                 float32 // geoid separation, ft, MSL minus HAE (used in altitude calculation)
	Satellites               uint16  // satellites used in solution
	SatellitesTracked        uint16  // satellites tracked (almanac data received)
	SatellitesSeen           uint16  // satellites seen (signal received)
	Accuracy                 float32 // 95% confidence for horizontal position, meters.
	NACp                     uint8   // NACp categories are defined in AC 20-165A
	Alt                      float32 // Feet MSL
	AccuracyVert             float32 // 95% confidence for vertical position, meters
	GPSVertVel               float32 // GPS vertical velocity, feet per second
	LastFixLocalTime         time.Time
	TrueCourse               float32
	GroundSpeed              uint16
	LastGroundTrackTime      time.Time
	GPSTime                  time.Time
	LastGPSTimeTime          time.Time // stratuxClock time since last GPS time received.
	LastValidNMEAMessageTime time.Time // time valid NMEA message last seen
	LastValidNMEAMessage     string    // last NMEA message processed.

	// From BMP180 pressure sensor.
	Temp              float64
	Pressure_alt      float64
	LastTempPressTime time.Time

	// From MPU6050 accel/gyro.
	Pitch            float64
	Roll             float64
	Gyro_heading     float64
	LastAttitudeTime time.Time
}

var rb *RockBLOCK.RockBLOCKSerialConnection

var mySituation SituationData

func situationGetter() {
	situationTicker := time.NewTicker(5 * time.Second)
	for {
		<-situationTicker.C
		url := "http://localhost/getSituation"
		resp, err := http.Get(url)
		if err != nil || !strings.HasPrefix(resp.Status, "200") {
			fmt.Printf("get situation error: %s\n", err.Error())
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("situation read err: %s\n", err.Error())
			continue
		}
		err = json.Unmarshal([]byte(body), &mySituation)
	}
}

func main() {
	r, err := RockBLOCK.NewRockBLOCKSerial()
	if err != nil {
		fmt.Printf("init error: %s\n", err.Error())
		return
	} else {
		fmt.Printf("initialized\n")
	}

	rb = r

	go situationGetter()

	sendTicker := time.NewTicker(2 * time.Minute)

	for {
		<-sendTicker.C
		t, err := r.GetTime()
		if err != nil {
			fmt.Printf("time error: %s\n", err.Error())
			continue
		} else {
			fmt.Printf("%s\n", t)
		}

		tB, _ := t.MarshalText()
		tS := string(tB)

		msg := fmt.Sprintf("%s,%0.4f,%0.4f", tS, mySituation.Lat, mySituation.Lng)
		fmt.Printf("msg=%s | len=%d. sending\n", msg, len(msg))

		rb.SendBinaryPersistent([]byte(msg))

	}

}
