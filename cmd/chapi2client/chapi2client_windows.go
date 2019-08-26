// (c) Copyright 2019 Hewlett Packard Enterprise Development LP

package main

import (
	"fmt"

	"github.com/hpe-storage/common-host-libs/chapi2/chapiclient"
	log "github.com/hpe-storage/common-host-libs/logger"
	"github.com/hpe-storage/common-host-libs/windows"
)

// getLogPath returns the folder where this application's log file will be stored
func getLogPath() string {
	return windows.LogPath
}

// getChapiClient allocates and returns a new CHAPI2 Client access object
func getChapiClient() (chapiClient *chapiclient.Client, err error) {
	if chapiClient, err = chapiclient.NewChapiWindowsClient("", nil); err != nil {
		log.Errorf("Failed getChapiClient, err=%v", err)
		fmt.Println(`Unable to get CHAPI client object.  Run this executable in the same folder as
the CHAPI server.  You can alternatively run "chapid.exe debug" to launch
chapid as an executable process instead of as a Windows service.`)
		fmt.Printf("\nError: %v\n", err)
		return nil, err
	}
	return chapiClient, err
}
