package main

import (
	"time"
)

const (
	REQUEST_NIL   = iota // No response required, just a status update.
	REQUEST_METAR        // Request a METAR response. Data is a field identifier.
)

type RockBLOCKTime struct {
	Time time.Time
}

func (t *RockBLOCKTime) UnmarshalText(text []byte) error {
	timeFormat := "02-01-06 15:04:05"
	t2, err := time.Parse(timeFormat, string(text))
	if err != nil {
		return err
	}
	*t = RockBLOCKTime{t2}
	return nil
}

type RockBLOCKIncoming struct {
	IMEI         string    `form:"imei"`
	MOMSN        int       `form:"momsn"`
	TransmitTime time.Time `form:"transmit_time"`
	IridiumLat   float64   `form:"iridium_latitude"`
	IridiumLng   float64   `form:"iridium_longitude"`
	IridiumCEP   float64   `form:"iridium_cep"`
	Data         []byte    `form:"data"`
}

// After decoding the 50 bytes.
type IridiumMessage struct {
	LatLngPresent bool
	Lat           float64
	Lng           float64
	RequestType   int
	Data          []byte
}

func (m *RockBLOCKIncoming) Process() IridiumMessage {

}
