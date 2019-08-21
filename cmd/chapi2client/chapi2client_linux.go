// (c) Copyright 2019 Hewlett Packard Enterprise Development LP

package main

import (
	"github.com/hpe-storage/common-host-libs/chapi2/chapiclient"
	log "github.com/hpe-storage/common-host-libs/logger"
)

// getLogPath returns the folder where this application's log file will be stored
func getLogPath() string {
	// TODO
	log.Fatal("getLogPath not implemented")
	return ""
}

func getChapiClient() (chapiClient *chapiclient.Client, err error) {
	// TODO
	log.Fatal("getChapiClient not implemented")
	return nil, nil
}
