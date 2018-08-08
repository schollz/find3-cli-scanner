package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/cihub/seelog"
)

func HopChannels(delay time.Duration) {
	log.Infof("hoping channels with %s delay", delay)
	// get list of channels
	stdout, _ := RunCommand(10*time.Second, "iwlist "+wifiInterface+" freq")
	channels := []string{}
	for _, line := range strings.Split(stdout, "\n") {
		if !strings.Contains(line, "Channel ") {
			continue
		}
		fs := strings.Fields(line)
		if len(fs) != 5 {
			continue
		}
		if fs[2] != ":" {
			continue
		}
		channels = append(channels, strings.Fields(line)[1])
	}
	log.Debugf("found %d channels: %+v", len(channels), channels)

	for {
		for _, channel := range channels {
			log.Debugf("switching to channel %s", channel)
			stdout, stderr := RunCommand(3*time.Second, fmt.Sprintf("iwconfig %s channel %s", wifiInterface, channel))
			currentChannel = channel
			log.Debug("output", stdout, stderr)
			time.Sleep(delay)
		}
	}
}
