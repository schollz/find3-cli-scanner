// +build !windows

package main

import (
	"errors"
	"sort"
	"strings"
	"time"

	log "github.com/cihub/seelog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/schollz/find3/server/main/src/models"
)

// Packet is the struct for the reverse scanning
type Packet struct {
	Mac       string    `json:"mac"`
	RSSI      int       `json:"rssi"`
	Timestamp time.Time `json:"timestamp"`
}

func ReverseScan(scanTime time.Duration) (sensors models.SensorData, err error) {
	log.Debugf("reverse scanning for %s", scanTime)
	sensors = models.SensorData{}
	sensors.Family = family
	sensors.Device = device
	sensors.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	sensors.Sensors = make(map[string]map[string]interface{})

	// make a channel to communicate timing
	done := make(chan error)

	go func() {
		log.Debug("waiting for ", scanTime)
		time.Sleep(scanTime)
		log.Debug("timed out")
		done <- nil
	}()

	packets := []Packet{}

	go func() {
		// gather packet information
		// Open device
		handle, err := pcap.OpenLive(wifiInterface, 2048, false, pcap.BlockForever)
		if err != nil {
			return
		}
		defer handle.Close()
		startTime := time.Now()

		// Use the handle as a packet source to process all packets
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			if time.Since(startTime).Seconds() > scanTime.Seconds() {
				done <- nil
				return
			}
			// Process packet here
			// fmt.Println(packet.String())
			address := ""
			rssi := 0
			for _, layer := range packet.Layers() {
				switch layer.LayerType() {
				case layers.LayerTypeRadioTap:
					rt := layer.(*layers.RadioTap)
					rssi = int(rt.DBMAntennaSignal)
				case layers.LayerTypeDot11:
					dot11 := layer.(*layers.Dot11)
					receiver := dot11.Address1.String()
					transmitter := dot11.Address2.String()
					if doAllPackets || receiver == "ff:ff:ff:ff:ff:ff" {
						address = transmitter
					}
				}
			}
			if address != "" && rssi != 0 {
				newPacket := Packet{
					Mac:       address,
					RSSI:      rssi,
					Timestamp: time.Now(),
				}
				packets = append(packets, newPacket)
				log.Debugf("%s: %d", newPacket.Mac, newPacket.RSSI)
			}
		}
		done <- err
	}()

	err = <-done
	log.Debug("got done signal")
	log.Debug(err)
	// merge packets
	strengths := make(map[string][]int)
	for _, packet := range packets {
		if _, ok := strengths[packet.Mac]; !ok {
			strengths[packet.Mac] = []int{}
		}
		strengths[packet.Mac] = append(strengths[packet.Mac], packet.RSSI)
	}
	mergedPackets := make(map[string]struct{})
	newPackets := make([]Packet, len(packets))
	i := 0
	for _, packet := range packets {
		if _, ok := mergedPackets[packet.Mac]; ok {
			continue
		}
		// get median value
		sort.Ints(strengths[packet.Mac])
		if len(strengths[packet.Mac]) > 2 {
			packet.RSSI = strengths[packet.Mac][len(strengths[packet.Mac])/2]
		} else {
			packet.RSSI = strengths[packet.Mac][0]
		}
		newPackets[i] = packet
		i++
		mergedPackets[packet.Mac] = struct{}{}
	}
	packets = newPackets[:i]
	log.Infof("collected %d packets", len(packets))
	if len(packets) == 0 {
		err = errors.New("no packets found")
	}
	sensors.Sensors["wifi"] = make(map[string]interface{})
	for _, packet := range packets {
		sensors.Sensors["wifi"][packet.Mac] = packet.RSSI
	}
	return
}

func PromiscuousMode(on bool) {
	sequence := []string{"ifconfig XX down", "iwconfig XX mode YY", "ifconfig XX up"}
	for _, command := range sequence {
		commandString := strings.Replace(command, "XX", wifiInterface, 1)
		if on {
			commandString = strings.Replace(commandString, "YY", "monitor", 1)
		} else {
			commandString = strings.Replace(commandString, "YY", "managed", 1)
		}
		s, t := RunCommand(60*time.Second, commandString)
		time.Sleep(1 * time.Second)
		if len(s) > 0 {
			log.Debugf("out: %s", s)
		}
		if len(t) > 0 {
			log.Debugf("err: %s", t)
		}
	}
}
