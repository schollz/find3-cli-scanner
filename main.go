package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/montanaflynn/stats"
	"github.com/schollz/find3/server/main/src/models"
	"github.com/urfave/cli"
)

var (
	wifiInterface string
	version       string
	commit        string
	date          string

	server                   string
	family, device, location string

	scanSeconds            int
	minimumThreshold       int
	doBluetooth            bool
	doWifi                 bool
	doReverse              bool
	doDebug                bool
	doGPS                  bool
	doSetPromiscuous       bool
	doNotModifyPromiscuity bool
	doIgnoreRandomizedMacs bool
	runForever             bool
)

func main() {
	defer log.Flush()
	app := cli.NewApp()
	app.Name = "find3-cli-scanner"
	if len(commit) > 6 {
		commit = commit[:6]
	}
	app.Version = fmt.Sprintf("%s (%s %s)", version, commit, date)
	app.Usage = "this command line scanner works with FIND3\n\t\tto capture bluetooth and WiFi signals from devices"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Zack Scholl",
			Email: "zack.scholl@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "bluetooth",
			Usage: "scan bluetooth",
		},
		cli.BoolFlag{
			Name:  "wifi",
			Usage: "scan wifi",
		},
		cli.StringFlag{
			Name:  "server",
			Value: "https://cloud.internalpositioning.com",
			Usage: "FIND3 server for submitting fingerprints",
		},
		cli.StringFlag{
			Name:  "interface,i",
			Value: "wlan0",
			Usage: "wifi interface for scanning",
		},
		cli.StringFlag{
			Name:  "family,f",
			Value: "",
			Usage: "family name",
		},
		cli.StringFlag{
			Name:  "device,d",
			Value: "",
			Usage: "device name",
		},
		cli.StringFlag{
			Name:  "location,l",
			Value: "",
			Usage: "location name (automatically toggles learning)",
		},
		cli.BoolFlag{
			Name:  "gps",
			Usage: "enable gps collection (using wifi)",
		},
		cli.BoolFlag{
			Name:  "passive",
			Usage: "enable passive scanning",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug mode",
		},
		cli.BoolFlag{
			Name:  "monitor-mode",
			Usage: "enable monitor mode (turn promiscuous mode on)",
		},
		cli.BoolFlag{
			Name:  "disable-monitor-mode",
			Usage: "disable monitor mode (turn promiscuous mode off)",
		},
		cli.BoolFlag{
			Name:  "no-modify",
			Usage: "disable changing wifi promiscuity mode",
		},
		cli.BoolFlag{
			Name:  "no-randomized-macs",
			Usage: "ignore randomized MAC addresses",
		},
		cli.BoolFlag{
			Name:  "forever",
			Usage: "run until Ctl+C signal",
		},
		cli.IntFlag{
			Name:  "min-rssi",
			Value: -100,
			Usage: "minimum RSSI to use",
		},
		cli.IntFlag{
			Name:  "scantime,s",
			Value: 40,
			Usage: "number of seconds to scan",
		},
	}
	app.Action = func(c *cli.Context) (err error) {
		// set variables
		server = c.GlobalString("server")
		family = c.GlobalString("family")
		device = c.GlobalString("device")
		wifiInterface = c.GlobalString("interface")
		location = c.GlobalString("location")
		doBluetooth = c.GlobalBool("bluetooth")
		doWifi = c.GlobalBool("wifi")
		doReverse = c.GlobalBool("passive")
		doDebug = c.GlobalBool("debug")
		doGPS = c.GlobalBool("gps")
		doSetPromiscuous = c.GlobalBool("monitor-mode")
		doNotModifyPromiscuity = c.GlobalBool("no-modify")
		doIgnoreRandomizedMacs = c.GlobalBool("no-randomized-macs")
		runForever = c.GlobalBool("forever")
		scanSeconds = c.GlobalInt("scantime")
		minimumThreshold = c.GlobalInt("min-rssi")

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

		if device == "" {
			return errors.New("device cannot be blank (set with -d)")
		} else if family == "" {
			return errors.New("family cannot be blank (set with -f)")
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
		return
	}
	err := app.Run(os.Args)
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

	if _, ok := payload.Sensors["wifi"]; ok && doGPS {
		acquired := 0.0
		for device := range payload.Sensors["wifi"] {
			lat, lon := func() (lat, lon float64) {
				type MacData struct {
					Ready      bool    `json:"ready"`
					MacAddress string  `json:"mac"`
					Exists     bool    `json:"exists"`
					Latitude   float64 `json:"lat,omitempty"`
					Longitude  float64 `json:"lon,omitempty"`
					Error      string  `json:"err,omitempty"`
				}
				var md MacData
				resp, err := http.Get("https://mac2gps.schollz.com/" + device)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				err = json.NewDecoder(resp.Body).Decode(&md)
				if err != nil {
					return
				}
				lat = md.Latitude
				lon = md.Longitude
				if md.Ready && md.Exists {
					log.Debugf("found GPS: %+v", md)
				}
				return
			}()
			if lat != 0 {
				acquired++
			}
			payload.GPS.Latitude += lat
			payload.GPS.Longitude += lon
		}
		if acquired > 0 {
			payload.GPS.Latitude = payload.GPS.Latitude / acquired
			payload.GPS.Longitude = payload.GPS.Longitude / acquired
		}
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
