package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// /metar/{IDENT}

func handleMETARRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	x := strings.Split(path, "/")
	if len(x) < 2 {
		http.Error(w, "Bad request.", 400)
		return
	}
	metar, err := getLatestADDSMETAR(x[1])
	if err == nil {
		w.Write([]byte(metar.Text + "\n"))
	}
}

// /receiveRockBLOCK

func handleRockBLOCKMsg(w http.ResponseWriter, r *http.Request) {

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
