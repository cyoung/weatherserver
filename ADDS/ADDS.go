package ADDS

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type ADDSTime struct {
	Time time.Time
}

func (t *ADDSTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	timeFormat := "2006-01-02T15:04:05Z"
	var inTime string
	d.DecodeElement(&inTime, &start)
	t2, err := time.Parse(timeFormat, inTime)
	if err != nil {
		return err
	}
	*t = ADDSTime{t2}
	return nil
}

type ADDSMETAR struct {
	Text           string   `xml:"raw_text"`
	StationID      string   `xml:"station_id"`
	Observation    ADDSTime `xml:"observation_time"`
	Latitude       float64  `xml:"latitude"`
	Longitude      float64  `xml:"longitude"`
	Temp           float64  `xml:"temp_c"`
	Dewpoint       float64  `xml:"dewpoint_c"`
	WindDirection  float64  `xml:"wind_dir_degrees"`
	WindSpeed      float64  `xml:"wind_speed_kt"`
	Visibility     float64  `xml:"visibility_statute_mi"`
	Altimeter      float64  `xml:"altim_in_hg"`
	FlightCategory string   `xml:"flight_category"`
}

type ADDSData struct {
	METARs []ADDSMETAR `xml:"METAR"`
}

type ADDSResponse struct {
	RequestIndex int      `xml:"request_index"`
	Data         ADDSData `xml:"data"`
}

func getADDSMETAR(ident string) ([]ADDSMETAR, error) {
	var ret ADDSResponse
	url := fmt.Sprintf("https://aviationweather.gov/adds/dataserver_current/httpparam?dataSource=metars&requestType=retrieve&format=xml&stationString=%s&hoursBeforeNow=2", ident)
	resp, err := http.Get(url)
	if err != nil || !strings.HasPrefix(resp.Status, "200") {
		return ret.Data.METARs, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret.Data.METARs, err
	}

	// Parse 'body'.
	err = xml.Unmarshal([]byte(body), &ret)

	return ret.Data.METARs, nil
}

func getLatestADDSMETAR(ident string) (ret ADDSMETAR, err error) {
	metars, errn := getADDSMETAR(ident)
	if errn != nil {
		return ret, errn
	}

	// Get the latest observation time.
	for _, v := range metars {
		if v.Observation.Time.After(ret.Observation.Time) {
			// This observation is later than the current one.
			ret = v
		}
	}

	return
}
