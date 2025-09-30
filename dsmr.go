package dsmr5

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"

	"go.bug.st/serial"
)

var ErrInvalidChecksum = errors.New("invalid checksum")

// Reads a raw telegram from the provided reader.
// Does not verify the CRC.
// The returned slice is only valid until the next ReadRawTelegram call.
func ReadRawTelegram(reader *bufio.Reader) ([]byte, error) {
	var bb bytes.Buffer
	inTelegram := false
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		log.Printf("read line: %q", line)

		trimmedLine := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmedLine, []byte("/")) {
			bb.Reset()
			bb.Write(line)
			inTelegram = true
			continue
		}

		if !inTelegram {
			continue
		}

		bb.Write(line)
		if bytes.HasPrefix(trimmedLine, []byte("!")) {
			return bb.Bytes(), nil
		}
	}
}

type SerialReader struct {
	port serial.Port
	r    *bufio.Reader
}

func NewSerialReader(portName string) (*SerialReader, error) {
	port, err := serial.Open(portName, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	})
	if err != nil {
		return nil, fmt.Errorf("error opening serial port %s: %w", portName, err)
	}
	r := bufio.NewReader(port)
	return &SerialReader{r: r, port: port}, nil
}

func (sr *SerialReader) Read() ([]byte, error) {
	content, err := ReadRawTelegram(sr.r)
	if err != nil {
		return nil, err
	}
	exci := bytes.LastIndex(content, []byte("!"))
	if exci == -1 || len(content) < exci+5 {
		return nil, errors.New("missing or invalid checksum marker")
	}
	crc16, err := strconv.ParseUint(string(content[exci:exci+5]), 16, 16)
	if err != nil {
		return nil, fmt.Errorf("illegal checksum: %w", err)
	}
	contentCRC16 := CRC16(content)
	if uint16(crc16) != contentCRC16 {
		return nil, fmt.Errorf("invalid checksum: expected %04X, got %04X", contentCRC16, crc16)
	}
	return content, nil
}

func (sr *SerialReader) Close() error {
	return sr.port.Close()
}
