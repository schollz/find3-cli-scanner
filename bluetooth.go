// +build !windows

package main

import (
	"strings"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/schollz/gatt"
)

var bdata map[string]map[string]interface{}
var bdatasync sync.Mutex

func onStateChanged(d gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		log.Debug("gatt powered on")
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	bdatasync.Lock()
	defer bdatasync.Unlock()
	bdata["bluetooth"][strings.ToLower(p.ID())] = rssi
}

var d gatt.Device
var bluetoothInitiated bool

func scanBluetooth(out chan map[string]map[string]interface{}) (err error) {
	log.Debug("scanning bluetooth")
	bdatasync.Lock()
	bdata = make(map[string]map[string]interface{})
	bdata["bluetooth"] = make(map[string]interface{})
	bdatasync.Unlock()

	if !bluetoothInitiated {
		log.Debug("initiating bluetooth")
		d, err = gatt.NewDevice()
		if err != nil {
			log.Debugf("Failed to open device, err: %s", err.Error())
			out <- bdata
			return
		}
		d.Handle(gatt.PeripheralDiscovered(onPeriphDiscovered))
		d.Init(onStateChanged)
		bluetoothInitiated = true
	}

	d.Scan([]gatt.UUID{}, false)
	select {
	case <-time.After(time.Duration(scanSeconds) * time.Second):
		log.Debug("bluetooth scan finished")
	}
	d.StopScanning()

	out <- bdata
	return
}
