package main

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	"github.com/paypal/gatt"
)

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
	fmt.Printf("\nPeripheral ID: %s %d\n", p.ID(), rssi)
}

func scanBluetooth(out chan map[string]map[string]interface{}) {
	log.Info("scanning bluetooth")

	data := make(map[string]map[string]interface{})
	data["bluetooth"] = make(map[string]interface{})

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
