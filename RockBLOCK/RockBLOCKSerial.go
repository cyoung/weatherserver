package RockBLOCK

import (
	"fmt"
	"github.com/tarm/serial"
)

type RockBLOCKSerialConnection struct {
	SerialConfig *serial.Config
	SerialPort   *serial.Port
	SerialIn     chan []byte
	SerialOut    chan []byte
}

func NewRockBLOCKSerial() (r *RockBLOCKSerialConnection, err error) {
	r = new(RockBLOCKSerialConnection)

	// Open serial port.
	cnf := &serial.Config{Name: "/dev/ttyAMA0", Baud: "19200"}
	p, errn := serial.OpenPort(serialConfig)
	if errn != nil {
		err := fmt.Errorf("serial port err: %s\n", errn.Error())
		return
	}

	// Serial port opened successfully.
	r.SerialConfig = cnf
	r.SerialPort = p

	// Initialize the device. If there's an error, return it.
	err = r.Init()

	return
}

func RockBLOCKScanSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We have a full \r-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

func (r *RockBLOCKSerialConnection) serialReader() {
	scanner := bufio.NewScanner(r.SerialPort)
	scanner.Split(RockBLOCKScanSplit)
	for scanner.Scan() {
		m := scanner.Bytes()
		r.SerialIn <- m
	}
}

func (r *RockBLOCKSerialConnection) serialWriter() {
	for {
		m := <-r.SerialOut
		err := r.SerialPort.Write(m)
		if err != nil {
			fmt.Printf("serial write error: %s\n", err.Error())
		}
	}
}

func (r *RockBLOCKSerialConnection) Init() error {
	// Set up the read/write channels.
	r.SerialIn = make([]byte)
	r.SerialOut = make([]byte)

	// Start the read/write goroutines.
	go r.serialReader()
	go r.serialWriter()

	// Send init command.
	r.SerialOut <- []byte("AT\r")

}
