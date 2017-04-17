package main

import (
	"../ADDS"
	"./RockBLOCK"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strings"
	"time"
)

var db *sql.DB

// /metar/{IDENT}

func handleMETARRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	x := strings.Split(path, "/")
	if len(x) < 2 {
		http.Error(w, "Bad request.", http.StatusBadRequest)
		return
	}
	metar, err := ADDS.GetLatestADDSMETARs(x[1])
	if err == nil {
		w.Write([]byte(metar.Text + "\n"))
	}
}

// /receiveRockBLOCK

func handleRockBLOCKMsg(w http.ResponseWriter, r *http.Request) {

	// Decode the form into 'RockBLOCKCOREIncoming'.
	var msg RockBLOCK.RockBLOCKCOREIncoming

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form could not be decoded", http.StatusBadRequest)
		return
	}

	// Ugly, but both gorilla/schema and ajg/form are unreliable here.
	if v, ok := r.Form["imei"]; ok {
		msg.IMEI = v[0]
	}
	if v, ok := r.Form["momsn"]; ok {
		msg.MOMSN = v[0]
	}
	if v, ok := r.Form["transmit_time"]; ok {
		msg.TransmitTime = v[0]
	}
	if v, ok := r.Form["iridium_latitude"]; ok {
		msg.IridiumLat = v[0]
	}
	if v, ok := r.Form["iridium_longitude"]; ok {
		msg.IridiumLng = v[0]
	}
	if v, ok := r.Form["iridium_cep"]; ok {
		msg.IridiumCEP = v[0]
	}
	if v, ok := r.Form["data"]; ok {
		d, err := hex.DecodeString(v[0])
		if err != nil {
			msg.Data = v[0]
		} else {
			msg.Data = string(d)
		}
	}

	// Process the message.
	msg.Process()

	x := strings.Split(msg.Data, ",")
	var TransmitInitTime time.Time
	var GPSLat string
	var GPSLng string

	if len(x) == 3 {
		TransmitInitTime.UnmarshalText([]byte(x[0]))
		GPSLat = x[1]
		GPSLng = x[2]
	}

	_, err := db.Exec(`INSERT INTO log SET IMEI=?, MOMSN=?, TransmitTime=?, IridiumLat=?, IridiumLng=?, IridiumCEP=?, InsertTime=NOW(), TransmitInitTime=?, GPSlat=?, GPSLng=?, Data=?`,
		msg.IMEI, msg.MOMSN, msg.TransmitTime, msg.IridiumLat, msg.IridiumLng, msg.IridiumCEP, TransmitInitTime, GPSLat, GPSLng, msg.Data)
	if err != nil {
		fmt.Printf("error inserting stats row to db: %s\n", err.Error())
	}

	// See if this is a plaintext weather request.
	if strings.HasPrefix(msg.Data, "METAR ") {
		x := strings.Split(msg.Data, " ")
		metar, err := ADDS.GetLatestADDSMETARs(x[1])
		if err == nil {
			m := new(RockBLOCK.RockBLOCKCOREOutgoing)
			m.IMEI = RockBLOCK.TEST_IMEI
			m.Data = []byte(metar.Text)
			if len(m.Data) > 50 {
				m.Data = m.Data[:50]
			}
			a, err := m.Send()
			fmt.Printf("attempt to send METAR %s\n", m.Data)
			if err != nil {
				fmt.Printf("a=%s, err=%s\n", a, err.Error())
			} else {
				fmt.Printf("a=%s\n", a)
			}
		}
	}

}

func main() {
	db2, err := sql.Open("mysql", "root:@/iridium")
	if err != nil {
		fmt.Printf("dbWriter(): db connect error: %s\n", err.Error())
		return
	}
	db = db2

	http.HandleFunc("/metar/", handleMETARRequest)
	http.HandleFunc("/receiveRockBLOCK", handleRockBLOCKMsg)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("managementInterface ListenAndServe: %s\n", err.Error())
	}
	for {
		time.Sleep(1 * time.Second)
	}
}
