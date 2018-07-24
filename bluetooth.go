package main

import (
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/paypal/gatt"
)

func scanBluetooth(out chan map[string]map[string]interface{}) {
	gatt.Debug = false
	log.Info("scanning bluetooth")

	data := make(map[string]map[string]interface{})
	data["bluetooth"] = make(map[string]interface{})

	onStateChanged := func(d gatt.Device, s gatt.State) {
		log.Debug("State:", s)
		switch s {
		case gatt.StatePoweredOn:
			log.Debug("scanning...")
			d.Scan([]gatt.UUID{}, false)
			return
		default:
			d.StopScanning()
		}
	}
	onPeriphDiscovered := func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
		data["bluetooth"][strings.ToLower(p.ID())] = rssi
	}

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

	out <- data
}
