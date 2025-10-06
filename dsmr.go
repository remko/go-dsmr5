package dsmr5

import (
	"bufio"
	"bytes"
	"fmt"
	"time"

	"go.bug.st/serial"
)

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
	if err := port.SetReadTimeout(3 * time.Second); err != nil {
		return nil, fmt.Errorf("error setting serial port timeout: %w", err)
	}
	r := bufio.NewReader(port)
	return &SerialReader{r: r, port: port}, nil
}

func (sr *SerialReader) Read() (*Telegram, error) {
	content, err := ReadRawTelegram(sr.r)
	if err != nil {
		return nil, err
	}
	return ParseTelegram(content)
}

func (sr *SerialReader) Close() error {
	return sr.port.Close()
}
