package dsmr5_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"mko.re/go/dsmr5"
)

//nolint:testableexamples,errcheck
func ExampleSerialReader() {
	r, err := dsmr5.NewSerialReader("/dev/ttyUSB0")
	if err != nil {
		panic(fmt.Errorf("error opening p1 port: %w", err))
	}
	defer r.Close()

	for {
		telegram, err := r.Read()
		if err != nil {
			log.Printf("error reading from serial: %v. re-attempting read in 1 second...\n", err)
			time.Sleep(1 * time.Second)
			continue
		}
		fmt.Printf("got telegram: %#v", telegram)
	}
}

// Fetch the latest Telegram from the HomeWizard P1 meter using the local API.
//
//nolint:noctx,testableexamples
func Example_homeWizard() {
	cl := &http.Client{Timeout: 5 * time.Second}
	resp, err := cl.Get(fmt.Sprintf("http://%s/api/v1/telegram", os.Getenv("HW_P1METER_HOST")))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	tb, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	t, err := dsmr5.ParseTelegram(tb)
	if err != nil {
		panic(err)
	}
	ci, err := t.CurrentImport()
	if err != nil {
		panic(err)
	}
	fmt.Printf("current power import: %v", ci)
}
