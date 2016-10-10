package RockBLOCK

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/tarm/serial"
	"strconv"
	"strings"
	"time"
)

var initTextMessage = []byte("AT+SBDWT=")
var initBinaryMessage = []byte("AT+SBDWB=")
var initSBDSessionExtended = []byte("AT+SBDIX")

type RockBLOCKSerialConnection struct {
	SerialConfig     *serial.Config
	SerialPort       *serial.Port
	SerialIn         chan []byte
	SerialOut        chan []byte
	processedBuffer  [][]byte
	ReceivedMessages []IridiumMessage
	SBDIX            SBDIXSerialResponse
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

/*
	parseSBDIX().
	 Parses a status response like:
	  +SBDIX: 0, 4, 1, 2, 6, 9
	 into a SBDIXSerialResponse structure, then saves it as 'SBDIX'.
*/

func (r *RockBLOCKSerialConnection) parseSBDIX(msg []byte) error {
	s := string(msg)
	if !strings.HasPrefix(s, "+SBDIX: ") {
		return errors.New("parseSBDIX(): Not a valid +SBDIX response.")
	}
	s = s[7:]
	x := strings.Split(s, ",")
	if len(x) != 6 {
		return errors.New("parseSBDIX(): Not a valid +SBDIX response.")
	}
	var parms []int
	for i := 0; i < len(x); i++ {
		c := strings.Trim(x[i], " ")
		i, err := strconv.ParseInt(c, 10, 32)
		if err != nil {
			return fmt.Errorf("parseSBDIX(): Not a valid +SBDIX response: %s.", s)
		}
		parms = append(parms, int(i))
	}

	r.SBDIX = SBDIXSerialResponse{
		MOStatus: parms[0],
		MOMSN:    parms[1],
		MTStatus: parms[2],
		MTMSN:    parms[3],
		MTLen:    parms[4],
		MTQueued: parms[5],
	}

	return nil
}

func (r *RockBLOCKSerialConnection) serialReader() {
	scanner := bufio.NewScanner(r.SerialPort)
	scanner.Split(RockBLOCKScanSplit)
	for scanner.Scan() {
		m := scanner.Bytes()
		m = bytes.Trim(m, "\r\n")
		if len(m) > 0 {
			// Automatic parsing.
			//TODO Parse all relevant information automatically.
			if StringPrefix(m, []byte("+SBDIX")) {
				r.parseSBDIX(m)
			}

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

type MsgEqualFunc func([]byte, []byte) bool

func StringEqual(a, b []byte) bool {
	return string(a) == string(b)
}

func StringPrefix(val, prefix []byte) bool {
	return strings.HasPrefix(string(val), string(prefix))
}

func (r *RockBLOCKSerialConnection) serialWait(comp []byte, eq MsgEqualFunc) error {
	timeoutTicker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case m := <-r.SerialIn:
			fmt.Printf("received: %s\n", string(m))
			r.processedBuffer = append(r.processedBuffer, m)
			if eq(m, comp) {
				return nil
			}
		case <-timeoutTicker.C:
			return errors.New("serialWait(): Timeout.")
		}
	}
	return errors.New("serialWait(): Unknown error.")
}

func (r *RockBLOCKSerialConnection) serialWaitEqual(s string) error {
	return r.serialWait([]byte(s), StringEqual)
}

func (r *RockBLOCKSerialConnection) serialWaitPrefix(prefix []byte) error {
	return r.serialWait(prefix, StringPrefix)
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
	err := r.serialWaitEqual("OK")
	if err != nil {
		return fmt.Errorf("init() error: %s", err.Error())
	}

	// Turn off flow control.
	r.serialWrite([]byte("AT&K0\r"))
	err = r.serialWaitEqual("OK")
	if err != nil {
		return fmt.Errorf("init() error: %s", err.Error())
	}

	return nil
}

func (r *RockBLOCKSerialConnection) SendText(msg []byte) error {
	cmd := append(initTextMessage, msg...)
	cmd = append(cmd, byte('\r'))
	r.serialWrite(cmd)
	err := r.serialWaitEqual("OK")
	if err != nil {
		return fmt.Errorf("SendText() error: %s", err.Error())
	}
	r.serialWrite(append(initSBDSessionExtended, byte('\r')))

	// Wait for "+SBDIX:" message
	return r.serialWaitPrefix([]byte("+SBDIX"))
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
	err := r.serialWaitEqual("READY")
	if err != nil {
		return fmt.Errorf("SendBinary() error: %s", err.Error())
	}

	msgWithChecksum := append(msg, r.binaryChecksum(msg)...)
	r.serialWrite(msgWithChecksum)

	// Wait for "0" (OK) response.
	err = r.serialWaitEqual("0")
	if err != nil {
		return fmt.Errorf("SendBinary() error: %s", err.Error())
	}

	r.serialWrite(append(initSBDSessionExtended, byte('\r')))

	// Wait for "+SBDIX:" message
	return r.serialWaitPrefix([]byte("+SBDIX"))

}
