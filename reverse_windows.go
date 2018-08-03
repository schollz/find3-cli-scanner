package main

import (
	"errors"
	"time"

	log "github.com/cihub/seelog"
	"github.com/schollz/find3/server/main/src/models"
)

func ReverseScan(scanTime time.Duration) (sensors models.SensorData, err error) {
	log.Debugf("reverse scanning for %s", scanTime)
	sensors = models.SensorData{}
	sensors.Family = family
	sensors.Device = device
	sensors.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	sensors.Sensors = make(map[string]map[string]interface{})
	err = errors.New("windows does not support gopacket")
	log.Error(err)
	return
}

func PromiscuousMode(on bool) {
	log.Error("windows does not support promiscuous mode")
}
