package main

// Copyright 2017 Nimble Storage, Inc.
import (
	"bytes"
	"encoding/json"
	log "github.com/hpe-storage/common-host-libs/logger"
	"github.com/hpe-storage/common-host-libs/model"
	"github.com/hpe-storage/common-host-libs/util"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	baseURL        = "http://localhost:9007/hosts"
	chapiClientLog = util.GetNltHome() + "log/chapiclient.log"
)

//TODO convert this to use testing framework
func main() {
	log.InitLogging(chapiClientLog, &log.LogParams{Level: "trace"}, false)

	var hosts model.Hosts
	var devices []*model.Device
	var device model.Device
	var serialnumber string
	var devicePartitions []*model.DevicePartition
	var buffer bytes.Buffer

	var restClient = &http.Client{
		Timeout: time.Second * 60,
	}

	// Retrieve the Host UUID
	response, err := restClient.Get(baseURL)
	if err != nil {
		return
	}
	defer response.Body.Close()
	buf, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(buf, &hosts)
	log.Tracef("URL:%s StatusCode:%d UUID:%s", baseURL, response.StatusCode, hosts[0].UUID)

	buffer.WriteString(baseURL)
	buffer.WriteString("/")
	buffer.WriteString(hosts[0].UUID)
	buffer.WriteString("/devices")
	devicesURL := buffer.String()
	log.Tracef("URL:%s StatusCode:%d", baseURL, response.StatusCode)

	// Retrive the Host Devices
	response, err = restClient.Get(devicesURL)
	if err != nil {
		return
	}
	defer response.Body.Close()
	buf, _ = ioutil.ReadAll(response.Body)
	json.Unmarshal(buf, &devices)
	for _, dev := range devices {
		log.Tracef("Device: %#v", dev)
		serialnumber = dev.SerialNumber
	}

	// Retrive the Device Info for a paritcular Device
	var buffDevice bytes.Buffer
	buffDevice.WriteString(devicesURL)
	buffDevice.WriteString("/")
	buffDevice.WriteString(serialnumber)
	deviceURL := buffDevice.String()
	response, err = restClient.Get(deviceURL)
	if err != nil {
		return
	}
	defer response.Body.Close()
	log.Tracef("URL:%s StatusCode:%d", baseURL, response.StatusCode)
	buf, _ = ioutil.ReadAll(response.Body)
	json.Unmarshal(buf, &device)
	log.Tracef("Device: %#v", device)

	// retrieve Device Partition For a Particular Device
	var buffPartition bytes.Buffer
	buffPartition.WriteString(deviceURL)
	buffPartition.WriteString("/")
	buffPartition.WriteString("partitions")
	partitionURL := buffPartition.String()
	response, err = restClient.Get(partitionURL)
	if err != nil {
		return
	}
	defer response.Body.Close()
	log.Tracef("URL:%s StatusCode:%d", baseURL, response.StatusCode)
	buf, _ = ioutil.ReadAll(response.Body)
	json.Unmarshal(buf, &devicePartitions)
	for _, part := range devicePartitions {
		log.Tracef("Partition: %#v", part)
	}
}
