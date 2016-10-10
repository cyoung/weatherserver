package RockBLOCK

const (
	MAX_MO_SZ = 340 // p.7 Iridium-9602-SBD-Transceiver-Product-Developers-Guide.pdf.
	MAX_MT_SZ = 270 // p.7 Iridium-9602-SBD-Transceiver-Product-Developers-Guide.pdf.
)

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
	MOStatus int // MO status provides an indication of the disposition of the mobile originated transaction.
	MOMSN    int // Mobile Originated Message Sequence Number.
	MTStatus int // MT status provides an indication of the disposition of the mobile terminated transaction.
	MTMSN    int // Mobile Terminated Message Sequence Number.
	MTLen    int // The length in bytes of the mobile terminated SBD message received from the GSS.
	MTQueued int // A count of mobile terminated SBD messages waiting at the GSS to be transferred to the device.
}

//TODO: Read 9602 outputs on:
//NetAv - network available. High = yes.
//RI - ring indicator. Low = ring.
