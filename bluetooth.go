package main

import (
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/schollz/gatt"
)

var bdata map[string]map[string]interface{}

func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	bdata[strings.ToLower(p.ID())] = rssi
}

func scanBluetooth(out chan map[string]map[string]interface{}) {
	log.Info("scanning bluetooth")

	bdata = make(map[string]map[string]interface{})
	bdata["bluetooth"] = make(map[string]interface{})

	d, err := gatt.NewDevice()
	if err != nil {
		log.Error("Failed to open device, err: %s\n", err)
		return
	}
	// Register handlers.
	d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
	d.Init(onStateChanged)
	select {
	case <-time.After(time.Duration(scanSeconds) * time.Second):
		log.Debug("bluetooth scan finished")
	}

	out <- bdata
}
