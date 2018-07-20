package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/montanaflynn/stats"
	"github.com/schollz/find3/server/main/src/models"
)

var (
	wifiInterface string

	server                   string
	family, device, location string

	scanSeconds            int
	minimumThreshold       int
	doBluetooth            bool
	doWifi                 bool
	doReverse              bool
	doDebug                bool
	doSetPromiscuous       bool
	doNotModifyPromiscuity bool
	runForever             bool
)

func main() {
	var err error
	defer log.Flush()
	flag.StringVar(&wifiInterface, "i", "wlan0", "wifi interface for scanning")
	flag.StringVar(&server, "server", "http://localhost:8003", "server to use")
	flag.StringVar(&family, "family", "", "family name")
	flag.StringVar(&device, "device", "", "device name")
	flag.StringVar(&location, "location", "", "location (optional)")
	flag.BoolVar(&doBluetooth, "bluetooth", false, "scan bluetooth")
	flag.BoolVar(&doWifi, "wifi", false, "scan wifi")
	flag.BoolVar(&doReverse, "passive", false, "passive scanning")
	flag.BoolVar(&doDebug, "debug", false, "enable debugging")
	flag.BoolVar(&doSetPromiscuous, "monitor-mode", false, "set promiscuous mode")
	flag.BoolVar(&doNotModifyPromiscuity, "no-modify", false, "disable changing wifi promiscuity mode")
	flag.BoolVar(&runForever, "forever", false, "run forever")
	flag.IntVar(&scanSeconds, "scantime", 40, "scan time")
	flag.IntVar(&minimumThreshold, "min-rssi", -100, "minimum RSSI to use")
	flag.Parse()

	if doDebug {
		setLogLevel("debug")
	} else {
		setLogLevel("info")
	}

	// ensure backwards compatibility
	if !doBluetooth && !doWifi {
		doWifi = true
	}

	if doSetPromiscuous {
		PromiscuousMode(true)
		return
	}

	if device == "" && doWifi {
		fmt.Println("device cannot be blank")
		flag.Usage()
		return
	}

	if family == "" {
		fmt.Println("family cannot be blank")
		flag.Usage()
		return
	}

	for {

		if doWifi {
			log.Infof("scanning with %s", wifiInterface)
		}
		if doBluetooth {
			log.Infof("scanning bluetooth")
		}
		if !doReverse {
			err = basicCapture()
		} else {
			log.Info("working in passive mode")
			err = reverseCapture()
		}
		if !runForever {
			break
		} else if err != nil {
			log.Warn(err)
		}
	}
	if err != nil {
		log.Error(err)
	}
}

func reverseCapture() (err error) {

	c := make(chan map[string]map[string]interface{})
	if doBluetooth {
		go scanBluetooth(c)
	}

	payload := models.SensorData{}
	payload.Family = family
	payload.Device = device
	payload.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	payload.Sensors = make(map[string]map[string]interface{})

	if doWifi {
		if !doNotModifyPromiscuity {
			PromiscuousMode(true)
			time.Sleep(1 * time.Second)
		}
		payload, err = ReverseScan(time.Duration(scanSeconds) * time.Second)
		if err != nil {
			return
		}
		if !doNotModifyPromiscuity {
			PromiscuousMode(false)
			time.Sleep(1 * time.Second)
		}
	}

	if doBluetooth {
		data := <-c
		log.Debugf("bluetooth data:%+v", data)
		for sensor := range data {
			payload.Sensors[sensor] = make(map[string]interface{})
			for device := range data[sensor] {
				payload.Sensors[sensor][device] = data[sensor][device]
			}
		}
	}
	bSensors, _ := json.MarshalIndent(payload, "", " ")
	log.Debug(string(bSensors))

	err = postData(payload, "/passive")
	return
}

func basicCapture() (err error) {
	payload := models.SensorData{}
	payload.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	payload.Family = family
	payload.Device = device
	payload.Location = location
	payload.Sensors = make(map[string]map[string]interface{})

	// collect sensors asynchronously
	c := make(chan map[string]map[string]interface{})
	numSensors := 0

	if doWifi {
		go iw(c)
		numSensors++
	}

	if doBluetooth {
		go scanBluetooth(c)
		numSensors++
	}

	for i := 0; i < numSensors; i++ {
		data := <-c
		for sensor := range data {
			payload.Sensors[sensor] = make(map[string]interface{})
			for device := range data[sensor] {
				payload.Sensors[sensor][device] = data[sensor][device]
			}
		}
	}

	if len(payload.Sensors) == 0 {
		err = errors.New("collected no data")
		return
	}
	bPayload, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		return
	}

	log.Debug(string(bPayload))
	err = postData(payload, "/data")
	return
}

// this doesn't work, just playing
func bluetoothTimeOfFlight() {
	t := time.Now()
	s, _ := RunCommand(60*time.Second, "l2ping -c 300 -f 0C:3E:9F:28:22:6A")
	milliseconds := make([]float64, 300)
	i := 0
	for _, line := range strings.Split(s, "\n") {
		if !strings.Contains(line, "ms") {
			continue
		}
		lineSplit := strings.Fields(line)
		msString := strings.TrimRight(lineSplit[len(lineSplit)-1], "ms")
		ms, err := strconv.ParseFloat(msString, 64)
		if err != nil {
			log.Error(err)
		}
		milliseconds[i] = ms
		i++
	}
	milliseconds = milliseconds[:i]
	median, err := stats.Median(milliseconds)
	if err != nil {
		log.Error(err)
	}
	fmt.Println(median)
	fmt.Println(time.Since(t) / 300)
}
