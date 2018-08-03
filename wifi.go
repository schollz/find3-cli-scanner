package main

import (
	log "github.com/cihub/seelog"
	"github.com/schollz/wifiscan"
)

func scanWifi(out chan map[string]map[string]interface{}) {
	datas := make(map[string]map[string]interface{})
	datas["wifi"] = make(map[string]interface{})
	wifis, err := wifiscan.Scan(wifiInterface)
	if err != nil {
		log.Error(err)
	}
	for _, w := range wifis {
		datas["wifi"][w.SSID] = w.RSSI
	}
	out <- datas
}
