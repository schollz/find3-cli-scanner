package main

import (
	"context"
	"regexp"
	"time"

	log "github.com/cihub/seelog"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/pkg/errors"
)

// sudo apt-get install bluez
// use btmgmt find instead
var negativeNumberRegex = regexp.MustCompile(`-\d+`)
var macAddressRegex = regexp.MustCompile(`([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`)

func scanBluetooth(out chan map[string]map[string]interface{}) {
	log.Info("scanning bluetooth")

	data := make(map[string]map[string]interface{})
	data["bluetooth"] = make(map[string]interface{})

	d, err := dev.NewDevice("default")
	if err != nil {
		log.Errorf("can't new device : %s", err)
		return
	}
	ble.SetDefaultDevice(d)

	// Scan for specified durantion, or until interrupted by user.
	log.Debugf("Scanning for %s...", time.Duration(scanSeconds)*time.Second)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), time.Duration(scanSeconds)*time.Second))
	devices := make(map[string][]float64)
	advHandler := func(a ble.Advertisement) {
		// fmt.Printf("[%s] %3d\n", a.Addr(), a.RSSI())
		if _, ok := devices[a.Addr().String()]; !ok {
			devices[a.Addr().String()] = []float64{}
		}
		devices[a.Addr().String()] = append(devices[a.Addr().String()], float64(a.RSSI()))
	}
	err = ble.Scan(ctx, true, advHandler, nil)
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		log.Debug("done")
	case context.Canceled:
		log.Debug("canceled\n")
	default:
		log.Error(err.Error())
	}

	for d := range devices {
		data["bluetooth"][d] = int(Average(devices[d]))
	}

	out <- data
}
