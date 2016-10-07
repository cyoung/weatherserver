package main

import (
	"./RockBLOCK"
	"fmt"
	"github.com/ajg/form"
	"net/http"
	"strings"
	"time"
)

// /metar/{IDENT}

func handleMETARRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	x := strings.Split(path, "/")
	if len(x) < 2 {
		http.Error(w, "Bad request.", http.StatusBadRequest)
		return
	}
	metar, err := getLatestADDSMETAR(x[1])
	if err == nil {
		w.Write([]byte(metar.Text + "\n"))
	}
}

// /receiveRockBLOCK

func handleRockBLOCKMsg(w http.ResponseWriter, r *http.Request) {

	// Decode the form into 'RockBLOCKCOREIncoming'.
	var msg RockBLOCKCOREIncoming

	d := form.NewDecoder(r.Body)

	if err := d.Decode(&u); err != nil {
		http.Error(w, "Form could not be decoded", http.StatusBadRequest)
		return
	}

	// Process the message.
	msg.Process()

}

func main() {
	http.HandleFunc("/metar/", handleMETARRequest)
	http.HandleFunc("/receiveRockBLOCK", handleRockBLOCKMsg)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("managementInterface ListenAndServe: %s\n", err.Error())
	}
	for {
		time.Sleep(1 * time.Second)
	}
}
