package main

import (
	"encoding/json"
	"fmt"
	"github.com/cyoung/ADDS"
	//	"github.com/cyoung/NEXRAD"
	"github.com/kellydunn/golang-geo"
	"github.com/stratux/goRFM95W/goRFM95W"
	"os"
	"sort"
	//	"strconv"
	"time"
)

type Config struct {
	StationLat          float64
	StationLng          float64
	StationServiceRange uint // Statute miles.
}

const (
	MAX_PACKET_SIZE = 255  // Bytes.
	MAX_PACKET_TIME = 1880 // ms. Calculated using the "LoRa Modem Calculator Tool", SF=12, BW=500kHz, CR=1, Payload=255, Preamble=4, CRC=Yes.
)

var myConfig Config

var selfGeo *geo.Point

var rfm95w *goRFM95W.RFM95W

func weatherUpdater() {
	updateTicker := time.NewTicker(5 * time.Minute)
	for {
		// Update the weather.
		//TODO: Need to add type-specific formatting that the correct meta-data for the report (lat/lng for PIREP, actually form coherent radar frames, etc.)
		// Get METARs.
		addsMetars, err := ADDS.GetLatestADDSMETARsInRadiusOf(myConfig.StationServiceRange, selfGeo)
		if err != nil {
			fmt.Printf("error obtaining METARs: %s\n", err.Error())
		} else {
			for _, metar := range addsMetars {
				// Generate a message, send it.
				m := DataMessage{
					Message:  []byte(metar.Text),
					UniqID:   "METAR " + metar.StationID,
					Priority: 10,
					Expiry:   time.Now().Add(15 * time.Minute),
				}
				messageChan <- m
			}
		}
		//FIXME: Only supporting METARs at the moment.
		/*
			// Get TAFs.
			addsTafs, err := ADDS.GetLatestADDSTAFsInRadiusOf(myConfig.StationServiceRange, selfGeo)
			if err != nil {
				fmt.Printf("error obtaining TAFs: %s\n", err.Error())
			} else {
				for _, taf := range addsTafs {
					// Generate a message, send it.
					m := DataMessage{
						Message:  []byte(taf.Text),
						UniqID:   "TAF " + taf.StationID,
						Priority: 11,
						Expiry:   time.Now().Add(15 * time.Minute),
					}
					messageChan <- m
				}
			}
			// Get PIREPs.
			addsPireps, err := ADDS.GetLatestADDSPIREPsInRadiusOf(myConfig.StationServiceRange, selfGeo)
			if err != nil {
				fmt.Printf("error obtaining PIREPs: %s\n", err.Error())
			} else {
				for _, pirep := range addsPireps {
					// Generate a message, send it.
					m := DataMessage{
						Message:  []byte(pirep.Text),
						UniqID:   "PIREP " + strconv.FormatFloat(pirep.Latitude, 'f', 5, 64) + "," + strconv.FormatFloat(pirep.Longitude, 'f', 5, 64),
						Priority: 9,
						Expiry:   time.Now().Add(15 * time.Minute),
					}
					messageChan <- m
				}
			}
			// Get NEXRAD.
			t, err := NEXRAD.GetCompressedTileFromLatLng(myConfig.StationLat, myConfig.StationLng)
			if err != nil {
				fmt.Printf("error obtaining NEXRAD: %s\n", err.Error())
			} else {
				m := DataMessage{
					Message:  t,
					UniqID:   "NEXRAD",
					Priority: 12,
					Expiry:   time.Now().Add(15 * time.Minute),
				}
				messageChan <- m
			}
		*/
		<-updateTicker.C
	}
}

type DataMessage struct {
	Message  []byte
	UniqID   string    // Identifier for the message. If another message is received with this same identifier, the new message replaces it.
	Priority int       // Priority is a non-unique. All messages of a single priority are grouped together, unordered.
	Expiry   time.Time // The message expires after this timestamp. It will not be sent after the maintenance period has passed and the sendList has been sent completely at least once.
}

var messageQueue map[string]DataMessage // UniqID -> DataMessage mapping.

func cleanupMessageQueue() {
	// Look for expired messages.
	t := time.Now()
	msgs := make(map[string]DataMessage, 0)
	for uniqID, msg := range messageQueue {
		if msg.Expiry.After(t) { // Not expired yet. Add to new queue.
			msgs[uniqID] = msg
		}
	}
	messageQueue = msgs // Copy over temporary queue.
}

/*
	makeSendList().
	 Orders the messageQueue by Priority, then creates chunks of size MAX_PACKET_SIZE.
*/

func makeSendList() [][]byte {
	ret := make([][]byte, 0)
	priorities := make([]int, len(messageQueue))
	var i int
	sendListWithPriorities := make(map[int][][]byte, 0)
	for _, msg := range messageQueue {
		sendListWithPriorities[msg.Priority] = append(sendListWithPriorities[msg.Priority], msg.Message)
		priorities[i] = msg.Priority
		i++
	}

	// Start creating packets of size MAX_PACKET_SIZE.
	sort.Ints(priorities)
	for i = 0; i < len(messageQueue); i++ {
		if msgs, ok := sendListWithPriorities[priorities[i]]; ok {
			for _, msg := range msgs {
				if len(msg) > MAX_PACKET_SIZE {
					//FIXME: Add provisions for fragmented packets.
					//					fmt.Printf("WARNING! Message is larger than max packet size: '%s'\n", string(msg))
					//					continue
					for len(msg) > MAX_PACKET_SIZE {
						ret = append(ret, msg[:MAX_PACKET_SIZE])
						msg = msg[MAX_PACKET_SIZE+1:]
					}
					ret = append(ret, msg)
				}
				if len(ret) > 0 && (len(ret[len(ret)-1])+len(msg)+1) < MAX_PACKET_SIZE {
					// Add this message to the last, with a '|' divider.
					ret[len(ret)-1] = append(ret[len(ret)-1], byte('|'))
					ret[len(ret)-1] = append(ret[len(ret)-1], msg...)
				} else {
					ret = append(ret, msg)
				}
			}
			delete(sendListWithPriorities, priorities[i]) // Remove this map key - we've finished with messages with this priority number.
		}
	}
	return ret
}

var messageChan chan DataMessage

func messageQueuer() {
	messageQueue = make(map[string]DataMessage, 0)

	var sendList [][]byte // Current message list.
	var sendPosition int  // Position in the sending list.
	var sendTimes int     // Number of times the current send list has been repeated.

	packetSenderTicker := time.NewTicker(MAX_PACKET_TIME * time.Millisecond)
	maintenanceTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case m := <-messageChan:
			// Receive a message to include in the next transmission.
			messageQueue[m.UniqID] = m
			fmt.Printf("Got message for '%s'!\n", m.UniqID)
		case <-packetSenderTicker.C:
			if len(sendList) == 0 {
				break // Nothing to send.
			}
			// Ready to send another packet. Send the next message in sendList.
			//			fmt.Printf("-->%s\n", string(sendList[sendPosition])) //TODO: Send message to LoRa transmitter.
			fmt.Printf("-->%d\n", len(sendList[sendPosition]))

			rfm95w.Send(sendList[sendPosition])

			sendPosition++
			if sendPosition+1 > len(sendList) {
				sendPosition = 0
				sendTimes++
			}
		case <-maintenanceTicker.C:
			// Do maintenance on the current queue. Clean up expired messages, create a new sendList, etc.
			if len(sendList) > 0 && sendTimes == 0 {
				// Don't do maintenance until the full list is sent at least once.
				fmt.Printf("Warning: Maintenance was triggered before sendList was sent completely. len(sendList)=%d, sendPosition=%d.\n", len(sendList), sendPosition)
				break
			}
			// Maintenance period has passed and the sendList has gone out at least once.
			// First delete the stale entries.
			cleanupMessageQueue()
			// Regenerate the send list.
			fmt.Printf("\n\n****Generating new send list****\n\n")
			sendList = makeSendList()
			// Print some statistics.
			numBytes := 0
			for _, m := range sendList {
				numBytes += len(m)
			}
			fmt.Printf("\nTotal sendList time=%dms, total bytes=%d, total packets=%d, packet efficiency=%.1f%%.\n", MAX_PACKET_TIME*len(sendList), numBytes, len(sendList), 100.0*float64(numBytes)/(float64(len(sendList)*MAX_PACKET_SIZE)))
			fmt.Printf("\n****Finished new send list****\n\n")
			// Re-set the send counters.
			sendPosition = 0
			sendTimes = 0
		}
	}
}

func main() {
	messageChan = make(chan DataMessage, 10240)

	// Read and parse config file.
	fp, err := os.Open("config.json")
	if err != nil {
		fmt.Printf("Can't open 'config.json'.\n")
		return
	}
	decoder := json.NewDecoder(fp)
	err = decoder.Decode(&myConfig)
	if err != nil {
		fmt.Printf("Couldn't read 'config.json'.\n")
		return
	}

	selfGeo = geo.NewPoint(myConfig.StationLat, myConfig.StationLng)

	// Initialize LoRa module with default values.
	rfm95w_h, err := goRFM95W.New(nil)
	if err != nil {
		fmt.Printf("LoRa: error: %s\n", err.Error())
		return
	} else {
		rfm95w = rfm95w_h
		// Start capturing.
		rfm95w.Start()
		fmt.Printf("LoRa module ready.\n")
	}

	go weatherUpdater()
	go messageQueuer()

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
