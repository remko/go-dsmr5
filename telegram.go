package dsmr5

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")

// OBIS code
type ObjectID string

type Telegram struct {
	Objects map[ObjectID][]AttributeValue
}

func (t Telegram) Time() (time.Time, error) {
	return t.timeValue("0-0:1.0.0", 0)
}

func (t Telegram) TotalImportTariff1() (float64, error) {
	return t.floatValue("1-0:1.8.1", 0)
}

func (t Telegram) TotalExportTariff1() (float64, error) {
	return t.floatValue("1-0:2.8.1", 0)
}

func (t Telegram) TotalImportTariff2() (float64, error) {
	return t.floatValue("1-0:1.8.2", 0)
}

func (t Telegram) TotalExportTariff2() (float64, error) {
	return t.floatValue("1-0:2.8.2", 0)
}

func (t Telegram) Peak() (float64, error) {
	return t.floatValue("1-0:1.4.0", 0)
}

func (t Telegram) MaxPeak() (float64, error) {
	return t.floatValue("1-0:1.6.0", 1)
}

func (t Telegram) MaxPeakTime() (time.Time, error) {
	return t.timeValue("1-0:1.6.0", 0)
}

func (t Telegram) CurrentImport() (float64, error) {
	return t.floatValue("1-0:1.7.0", 0)
}

func (t Telegram) CurrentExport() (float64, error) {
	return t.floatValue("1-0:2.7.0", 0)
}

func (t Telegram) TotalGasImport() (float64, error) {
	return t.floatValue("0-1:24.2.3", 1)
}

func (t Telegram) GasTime() (time.Time, error) {
	return t.timeValue("0-1:24.2.3", 0)
}

func (t Telegram) timeValue(id ObjectID, n int) (time.Time, error) {
	v, ok := t.Objects[id]
	if !ok || len(v) == 0 {
		return time.Time{}, ErrNotFound
	}
	return v[n].TimestampValue()
}

func (t Telegram) floatValue(id ObjectID, n int) (float64, error) {
	v, ok := t.Objects[id]
	if !ok || len(v) == 0 {
		return 0, ErrNotFound
	}
	return v[n].FloatValue()
}

type AttributeValue struct {
	Value string
	Unit  string
}

func (t AttributeValue) FloatValue() (float64, error) {
	return strconv.ParseFloat(t.Value, 64)
}

func (t AttributeValue) TimestampValue() (time.Time, error) {
	if len(t.Value) != 13 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", t.Value)
	}
	dateTimePart := t.Value[:12]

	var targetLoc *time.Location
	switch t.Value[12] {
	case 'S':
		targetLoc = time.FixedZone("CEST", 2*3600)
	case 'W':
		targetLoc = time.FixedZone("CET", 1*3600)
	default:
		return time.Time{}, fmt.Errorf("invalid timestamp timezone specifier: %s", t.Value)
	}
	tm, err := time.ParseInLocation("060102150405", dateTimePart, targetLoc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp: %w", err)
	}

	return tm, nil
}

var obisRE = regexp.MustCompile(`^(\d-\d:\d+\.\d+\.\d+)(.*)$`)
var valueRE = regexp.MustCompile(`\(([^)\*]+)(\*([^)]+))?\)`)

func ParseTelegram(data []byte) (*Telegram, error) {
	exci := bytes.LastIndex(data, []byte("!"))
	if exci == -1 || len(data) < exci+5 {
		return nil, errors.New("missing or invalid checksum marker")
	}
	crc16, err := strconv.ParseUint(string(data[exci+1:exci+5]), 16, 16)
	if err != nil {
		return nil, fmt.Errorf("illegal checksum: %w", err)
	}
	dataCRC16 := CRC16(data[:exci+1])
	if uint16(crc16) != dataCRC16 {
		return nil, fmt.Errorf("invalid checksum: expected %04X, got %04X", dataCRC16, crc16)
	}
	return parseTelegram(data)
}

func parseTelegram(data []byte) (*Telegram, error) {
	t := Telegram{Objects: make(map[ObjectID][]AttributeValue)}
	dataStarted := false
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "/") {
			dataStarted = true
			continue
		}
		if !dataStarted {
			continue
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "!") {
			break
		}
		m := obisRE.FindStringSubmatch(line)
		if len(m) != 3 {
			log.Printf("unrecognized line: %s", line)
			continue
		}
		values := []AttributeValue{}
		vm := valueRE.FindAllStringSubmatch(m[2], -1)
		for _, v := range vm {
			values = append(values, AttributeValue{Value: v[1], Unit: v[3]})
		}
		t.Objects[ObjectID(m[1])] = values
	}
	return &t, nil
}
