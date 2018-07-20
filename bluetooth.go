package main

import (
	"context"
	"regexp"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

// sudo apt-get install bluez
// use btmgmt find instead
var negativeNumberRegex = regexp.MustCompile(`-\d+`)
var macAddressRegex = regexp.MustCompile(`([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`)

func scanBluetooth(out chan map[string]map[string]interface{}) {
	var err error
	data := make(map[string]map[string]interface{})
	data["bluetooth"], err = getbluetoothScan()
	if err != nil {
		log.Error(err)
	}
	out <- data
}

func getbluetoothScan() (devices map[string]interface{}, err error) {
	devices = make(map[string]interface{})
	d, err := dev.NewDevice("default")
	if err != nil {
		return
	}
	defer d.Stop()

	ble.SetDefaultDevice(d)
	// Default to search device with name of Gopher (or specified by user).
	filter := func(a ble.Advertisement) {
		log.Debug(a.Addr(), a.RSSI())
		devices[a.Addr().String()] = a.RSSI()
	}

	sd := 5 * time.Second
	// Scan for specified durantion, or until interrupted by user.
	log.Debugf("Scanning for %s...", sd.String())
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), sd))
	errFinish := ble.Scan(ctx, false, filter, nil)
	if errFinish != nil {
		if !strings.Contains(errFinish.Error(), "context deadline") {
			err = errFinish
		}
	}
	return
}
