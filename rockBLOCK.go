package main

import (
	"time"
)

type RockBLOCKIncoming struct {
	IMEI         string
	MOMSN        int
	TransmitTime time.Time
	IridiumLat   float64
	IridiumLng   float64
	IridiumCEP   float64
	Data         []byte
}
