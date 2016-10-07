package RockBLOCK

// After decoding the 50 bytes.
type IridiumMessage struct {
	LatLngPresent bool
	Lat           float64
	Lng           float64
	RequestType   int
	Data          []byte
}

// Iridium modem state.
type SBDIXSerialResponse struct {
	MOStatus string // MO status provides an indication of the disposition of the mobile originated transaction.
	MOMSN    string // Mobile Originated Message Sequence Number.
	MTStatus string // MT status provides an indication of the disposition of the mobile terminated transaction.
	MTMSN    string // Mobile Terminated Message Sequence Number.
	MTLen    string // The length in bytes of the mobile terminated SBD message received from the GSS.
	MTQueued string // A count of mobile terminated SBD messages waiting at the GSS to be transferred to the device.
}
