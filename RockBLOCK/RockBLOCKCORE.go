package RockBLOCK

import (
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
	IMEI         string    `form:"imei"`
	MOMSN        int       `form:"momsn"`
	TransmitTime time.Time `form:"transmit_time"`
	IridiumLat   float64   `form:"iridium_latitude"`
	IridiumLng   float64   `form:"iridium_longitude"`
	IridiumCEP   float64   `form:"iridium_cep"`
	Data         []byte    `form:"data"`
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
	m.Username = "a"
	m.Password = "b"
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
