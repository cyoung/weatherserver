package RockBLOCK

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/tarm/serial"
	"time"
)

var initTextMessage = []byte("AT+SBDWT=")
var initBinaryMessage = []byte("AT+SBDWB=")

type RockBLOCKSerialConnection struct {
	SerialConfig     *serial.Config
	SerialPort       *serial.Port
	SerialIn         chan []byte
	SerialOut        chan []byte
	processedBuffer  [][]byte
	ReceivedMessages []IridiumMessage
}

func NewRockBLOCKSerial() (r *RockBLOCKSerialConnection, err error) {
	r = new(RockBLOCKSerialConnection)

	// Open serial port.
	cnf := &serial.Config{Name: "/dev/pts/4", Baud: 19200}
	p, errn := serial.OpenPort(cnf)
	if errn != nil {
		err = fmt.Errorf("serial port err: %s\n", errn.Error())
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
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func (r *RockBLOCKSerialConnection) serialReader() {
	scanner := bufio.NewScanner(r.SerialPort)
	scanner.Split(RockBLOCKScanSplit)
	for scanner.Scan() {
		m := scanner.Bytes()
		m = bytes.Trim(m, "\r\n")
		if len(m) > 0 {
			r.SerialIn <- bytes.Trim(m, "\r\n")
		}
	}
}

func (r *RockBLOCKSerialConnection) serialWriter() {
	for {
		m := <-r.SerialOut
		_, err := r.SerialPort.Write(m)
		if err != nil {
			fmt.Printf("serial write error: %s\n", err.Error())
		}
	}
}

func (r *RockBLOCKSerialConnection) serialWrite(m []byte) {
	r.SerialOut <- m
}

func (r *RockBLOCKSerialConnection) serialWait(s string) error {
	timeoutTicker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case m := <-r.SerialIn:
			fmt.Printf("received: %s\n", string(m))
			r.processedBuffer = append(r.processedBuffer, m)
			if string(m) == s {
				return nil
			}
		case <-timeoutTicker.C:
			return errors.New("serialWait(): Timeout.")
		}
	}
	return errors.New("serialWait(): Unknown error.")
}

func (r *RockBLOCKSerialConnection) Init() error {
	// Set up the read/write channels.
	r.SerialIn = make(chan []byte)
	r.SerialOut = make(chan []byte)

	// Start the read/write goroutines.
	go r.serialReader()
	go r.serialWriter()

	// Send init command.
	r.serialWrite([]byte("AT\r"))
	err := r.serialWait("OK")
	if err != nil {
		return fmt.Errorf("init() error: %s", err.Error())
	}

	// Turn off flow control.
	r.serialWrite([]byte("AT&K0\r"))
	err = r.serialWait("OK")
	if err != nil {
		return fmt.Errorf("init() error: %s", err.Error())
	}

	return nil
}

func (r *RockBLOCKSerialConnection) SendText(msg []byte) error {
	cmd := append(initTextMessage, msg...)
	cmd = append(cmd, byte('\r'))
	r.serialWrite(cmd)
	return r.serialWait("OK")
}

func (r *RockBLOCKSerialConnection) binaryChecksum(msg []byte) []byte {
	var sum int32
	for i := 0; i < len(msg); i++ {
		sum += int32(msg[i])
	}
	return []byte{byte((sum & 0xFF00) >> 8), byte(sum & 0xFF)}
}

func (r *RockBLOCKSerialConnection) SendBinary(msg []byte) error {
	msgLen := len(msg)
	cmd := append(initBinaryMessage, []byte(fmt.Sprintf("%d\r", msgLen))...)
	r.serialWrite(cmd)

	// Wait for the "READY" message, then send the whole binary message plus the checksum.
	err := r.serialWait("READY")
	if err != nil {
		return fmt.Errorf("SendBinary() error: %s", err.Error())
	}

	msgWithChecksum := append(msg, r.binaryChecksum(msg)...)
	r.serialWrite(msgWithChecksum)

	// Wait for "0" (OK) response.
	err = r.serialWait("0")
	if err != nil {
		return fmt.Errorf("SendBinary() error: %s", err.Error())
	}

	return nil

}
