package main

import log "github.com/cihub/seelog"

func scanBluetooth(out chan map[string]map[string]interface{}) {
	log.Error("windows has no bluetooth interface")
	bdata := make(map[string]map[string]interface{})
	bdata["bluetooth"] = make(map[string]interface{})
	out <- bdata
}
