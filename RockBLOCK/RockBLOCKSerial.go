//+SBDDET
//+SBDDSC
//+SBDREG
//+SBDAREG
//+SBDMTA

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
var initSBDSession = []byte("AT+SBDI")
var getSignalQualityMessage = []byte("AT+CSQ")
var downloadBinaryMessage = []byte("AT+SBDRT")
var requestSystemTimeMessage = []byte("AT-MSSTM")
var clearBuffers = []byte("AT+SBDD0")

type RockBLOCKSerialConnection struct {
	SerialConfig     *serial.Config
	SerialPort       *serial.Port
	SerialIn         chan []byte
	SerialOut        chan []byte
	processedBuffer  [][]byte
	ReceivedMessages []IridiumMessage
	SBDI             SBDISerialResponse
	SignalQuality    int
	SystemTime       time.Time
}

func NewRockBLOCKSerial() (r *RockBLOCKSerialConnection, err error) {
	r = new(RockBLOCKSerialConnection)

	// Open serial port.
	cnf := &serial.Config{Name: "/dev/ttyUSB0", Baud: 19200}
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
	parseSBDI().
	 Parses a status response like:
	  +SBDIX: 0, 4, 1, 2, 6, 9
	  +SBDI: 0, 4, 1, 2, 6, 9
	 into a SBDISerialResponse structure, then saves it as 'SBDI'.
*/

func (r *RockBLOCKSerialConnection) parseSBDI(msg []byte) error {
	s := string(msg)
	if !strings.HasPrefix(s, "+SBDI") {
		return errors.New("parseSBDI(): Not a valid +SBDI_ response.")
	}
	s = s[7:]
	x := strings.Split(s, ",")
	if len(x) != 6 {
		return errors.New("parseSBDI(): Not a valid +SBDI_ response.")
	}
	var parms []int
	for i := 0; i < len(x); i++ {
		c := strings.Trim(x[i], " ")
		i, err := strconv.ParseInt(c, 10, 32)
		if err != nil {
			return fmt.Errorf("parseSBDI(): Not a valid +SBDI_ response: %s.", s)
		}
		parms = append(parms, int(i))
	}

	r.SBDI = SBDISerialResponse{
		MOStatus: parms[0],
		MOMSN:    parms[1],
		MTStatus: parms[2],
		MTMSN:    parms[3],
		MTLen:    parms[4],
		MTQueued: parms[5],
	}

	return nil
}

func (r *RockBLOCKSerialConnection) parseCSQ(msg []byte) error {
	s := string(msg)
	if !strings.HasPrefix(s, "+CSQ:") {
		return errors.New("parseCSQ(): Not a valid +SBDI response.")
	}
	s = s[5:]
	v := strings.Trim(s, " ")
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return fmt.Errorf("parseCSQ(): Not a valid +CSQ response: %s.", s)
	}
	r.SignalQuality = int(i)
	return nil
}

func (r *RockBLOCKSerialConnection) parseMSSTM(msg []byte) error {
	s := string(msg)
	if !strings.HasPrefix(s, "-MSSTM:") {
		return errors.New("parseMSSTM(): Not a valid -MSSTM response.")
	}
	s = s[7:]
	v := strings.Trim(s, " ")
	i, err := strconv.ParseInt(v, 16, 32)
	if err != nil {
		return fmt.Errorf("parseMSSTM(): Not a valid -MSSTM response.")
	}

	// Era2: https://www.g1sat.com/download/iridium/2015%20Iridium%20Time%20Epoch%20Change%20ITN0018%20v1.2.pdf.
	iridiumEpochTime := time.Date(2014, 5, 11, 14, 23, 55, 0, time.UTC)

	// -MMSTM returns the number of 90ms intervals since Iridium Epoch, unless it has rolled over.
	//FIXME: Rollover detection.

	r.SystemTime = iridiumEpochTime.Add(90 * time.Millisecond * time.Duration(i))
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
			if StringPrefix(m, []byte("+SBDI")) {
				r.parseSBDI(m)
			}
			if StringPrefix(m, []byte("+CSQ:")) {
				r.parseCSQ(m)
			}
			if StringPrefix(m, []byte("-MSSTM:")) {
				r.parseMSSTM(m)
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
	fmt.Printf("sent: %s\n", string(m))
	r.SerialOut <- m
}

type MsgEqualFunc func([]byte, []byte) bool

func StringEqual(a, b []byte) bool {
	return string(a) == string(b)
}

func StringPrefix(val, prefix []byte) bool {
	return strings.HasPrefix(string(val), string(prefix))
}

// For parsed commands, the return value comes after it has been parsed.
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

func (r *RockBLOCKSerialConnection) clearBuffer() error {
	cmd := append(clearBuffers, byte('\r'))
	r.serialWrite(cmd)
	return r.serialWaitEqual("OK")
}

func (r *RockBLOCKSerialConnection) SendText(msg []byte) error {
	r.clearBuffer()
	cmd := append(initTextMessage, msg...)
	cmd = append(cmd, byte('\r'))
	r.serialWrite(cmd)
	err := r.serialWaitEqual("OK")
	if err != nil {
		return fmt.Errorf("SendText() error: %s", err.Error())
	}
	r.serialWrite(append(initSBDSession, byte('\r')))

	// Wait for "+SBDI:" message
	err = r.serialWaitPrefix([]byte("+SBDI:"))
	if err != nil {
		return err
	}

	err = r.serialWaitEqual("OK")
	if err != nil {
		return fmt.Errorf("SendText() error: %s", err.Error())
	}

	// Check if message was sent successfully.
	if r.SBDI.MOStatus != 2 {
		return fmt.Errorf("Send message error: %v", r.SBDI)
	}

	return nil

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

	r.serialWrite(append(initSBDSession, byte('\r')))

	// Wait for "+SBDI:" message
	err = r.serialWaitPrefix([]byte("+SBDI:"))
	if err != nil {
		return err
	}

	// Check if message was sent successfully.
	if r.SBDI.MOStatus != 1 {
		return fmt.Errorf("Send message error: %v", r.SBDI)
	}

	return nil

}

func (r *RockBLOCKSerialConnection) GetSignalQuality() (int, error) {
	msg := append(getSignalQualityMessage, byte('\r'))
	r.serialWrite(msg)
	if err := r.serialWaitPrefix([]byte("+CSQ:")); err != nil {
		return -1, err
	}

	return r.SignalQuality, nil

}

/*
	WaitForNetwork().
	 Returns nil if and only if a signal quality indicator greater than 0 is encountered in less than 't'.
	 Checks once per 5 seconds.
*/
func (r *RockBLOCKSerialConnection) WaitForNetwork(t time.Duration) error {
	finishTicker := time.NewTicker(t)
	checkTicker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-finishTicker.C:
			return errors.New("Timeout.")
		case <-checkTicker.C:
			signal, err := r.GetSignalQuality()
			if err != nil {
				return err
			}
			if signal != 0 {
				return nil
			}
		}
	}
	return errors.New("Timeout.")
}

func (r *RockBLOCKSerialConnection) DownloadMessage() error {
	// Check if we have messages waiting.
	if r.SBDI.MTStatus != 1 {
		// No messages.
		return errors.New("DownloadMessages(): No messages waiting.")
	}

	// Initiate the download.
	_ = append(downloadBinaryMessage, byte('\r'))

	//TODO: Switch to data mode, specified number of bytes.

	return nil
}

func (r *RockBLOCKSerialConnection) GetTime() (time.Time, error) {
	msg := append(requestSystemTimeMessage, byte('\r'))
	r.serialWrite(msg)
	if err := r.serialWaitPrefix([]byte("-MSSTM:")); err != nil {
		return time.Now(), err // time.Now(): Best effort.
	}

	r.serialWaitEqual("OK")

	return r.SystemTime, nil
}
