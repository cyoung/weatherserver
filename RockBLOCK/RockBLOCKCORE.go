package RockBLOCK

import (
	"encoding/hex"
	"errors"
	"github.com/ajg/form"
	"io/ioutil"
	"net/http"
	"strings"
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

type RockBLOCKCOREIncoming struct {
	IMEI         string `form:"imei"`
	MOMSN        string `form:"momsn"`
	TransmitTime string `form:"transmit_time"`
	IridiumLat   string `form:"iridium_latitude"`
	IridiumLng   string `form:"iridium_longitude"`
	IridiumCEP   string `form:"iridium_cep"`
	Data         string `form:"data"`
}

type RockBLOCKCOREOutgoing struct {
	IMEI     string `form:"imei"`
	Username string `form:"username"`
	Password string `form:"password"`
	Data     []byte `form:"data"`
}

func (m *RockBLOCKCOREIncoming) Process() IridiumMessage {
	//TODO.
	var ret IridiumMessage
	return ret
}

func (m *RockBLOCKCOREOutgoing) Send() (string, error) {
	m.Username = CORE_USER
	m.Password = CORE_PASS

	if len(m.IMEI) == 0 || len(m.Data) == 0 {
		return "", errors.New("Insufficient data.")
	}

	// Hex-encode the 'Data' value.
	encodedData := make([]byte, hex.EncodedLen(len(m.Data)))
	hex.Encode(encodedData, m.Data)
	m.Data = encodedData

	vals, err := form.EncodeToValues(m)
	if err != nil {
		return "", err
	}

	// Get the response.
	resp, err := http.PostForm("https://core.rock7.com/rockblock/MT", vals)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	x := strings.Split(string(body), ",")
	if x[0] == "OK" {
		// Success.
		return x[1], nil
	}

	// Is there a valid error response?
	if len(x) > 2 {
		return "", errors.New(strings.Join(x[1:], ","))
	}

	// Not even a valid error response.
	return "", errors.New("Invalid response.")
}
