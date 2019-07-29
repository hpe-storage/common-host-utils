//go:generate updatewinversioninfo
//go:generate goversioninfo -manifest=chapid.exe.manifest

// (c) Copyright 2019 Hewlett Packard Enterprise Development LP

// Copyright 2012 The Go Authors. All rights reserved.
// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hpe-storage/common-host-libs/chapi2"
	log "github.com/hpe-storage/common-host-libs/logger"
	"github.com/hpe-storage/common-host-libs/util"
	"github.com/hpe-storage/common-host-libs/windows"
	"github.com/hpe-storage/common-host-libs/windows/winservice"
	"github.com/hpe-storage/common-host-libs/windows/wmi"
	"golang.org/x/sys/windows/svc"
)

const (
	svcName            = "HPEVolumeService"
	svcDisplayName     = "HPE Volume Management Service"
	svcDescription     = "Provides services needed to manage HPE volumes"
	svcLogPathTemplate = "chapid-%v.log" // %v is the instance's location GUID; see NWT-3508 for details
)

var (
	// Version contains the current version added by the build process
	Version = "dev"
	// Commit containers the hg commit added by the build process
	Commit = "unknown"
	// Our CHAPI2 Windows service framework
	chapiService winservice.WinService
)

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, or stop.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func cleanup() {
	wmi.Cleanup()
}

func main() {
	// Configure our trace log file; start by enumerating the logs folder
	chapiLogFolder := windows.LogPath

	// We want instance specific CHAPI log files (NWT-3508); enumerate our log filename
	exePath, _ := os.Executable()
	exePath = filepath.Dir(exePath)
	hasher := md5.New()
	hasher.Write([]byte(strings.ToLower(exePath))) // Create hash from lower case path
	logFilename := fmt.Sprintf(svcLogPathTemplate, hex.EncodeToString(hasher.Sum(nil)))

	// Now create our log file
	chapiLogFile := filepath.Join(chapiLogFolder, logFilename)
	log.InitLogging(chapiLogFile, nil, false)
	defer cleanup()

	// Log the service name, any input parameters, and the log path
	log.Infoln(strings.Repeat("-", 80))
	log.Infof("%v, Version=%v, Commit=%v", svcDisplayName, Version, Commit)
	log.Infof("CLI: %v", strings.Join(os.Args, " "))
	log.Infof("LOG: %v", chapiLogFile)

	// Initialize our winservice.WinService object
	chapiService.UseEventLog = false
	chapiService.Start = serviceStart
	chapiService.Stop = serviceStop

	isIntSess, err := svc.IsAnInteractiveSession()
	log.Infof("IsAnInteractiveSession=%v", isIntSess)
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		chapiService.RunService(svcName, false)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		chapiService.RunService(svcName, true)
		return
	case "install":
		err = chapiService.InstallService(svcName, svcDisplayName, svcDescription)
	case "remove":
		err = chapiService.RemoveService(svcName)
	case "start":
		err = chapiService.StartService(svcName)
	case "stop":
		err = chapiService.ControlService(svcName, svc.Stop, svc.Stopped)
	case "pause":
		err = chapiService.ControlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		err = chapiService.ControlService(svcName, svc.Continue, svc.Running)
	case "wmi":
		// Perform WMI queries (internal use only)
		log.Infof("WMI query:  %v", strings.Join(os.Args[1:], " "))
		if len(os.Args) > 2 {
			wmiClass := os.Args[2]
			loopCount := 1
			if len(os.Args) > 3 {
				loopCount, err = strconv.Atoi(os.Args[3])
			}
			if err == nil {
				err = wmiQuery(wmiClass, loopCount)
			}
		}
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		util.LogError.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
	return
}

// serviceStart is called when the winservice framework starts the service
func serviceStart() {
	// Start CHAPI2
	chapi2.Run()
}

// serviceStop is called when the winservice framework stops the service
func serviceStop() {
	// Stop CHAPI2
	chapi2.StopChapid()
}

// wmiQuery is called from our CLI handler if the caller wants us to perform a WMI query
func wmiQuery(wmiClass string, loopCount int) (err error) {

	log.Infof(">>>>> wmiQuery called, wmiClass=%v, loopCount=%v", wmiClass, loopCount)
	defer log.Info("<<<<< wmiQuery")

	// WMI class name that causes all wrapped WMI classes to be enumerated
	const queryAll = "all"

	var wmiResult interface{}
	var wmiResults []interface{}

	for i := 0; (loopCount == 0) || (i < loopCount); {

		// Clear our results slice
		wmiResults = nil

		// wmi.GetMSFC_FibrePortHBAAttributes
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "msfc_fibreporthbaattributes") {
			wmiResult, err = wmi.GetMSFC_FibrePortHBAAttributes()
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetMSFTDisk
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "msft_disk") {
			wmiResult, err = wmi.GetMSFTDisk("")
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetMSFTPartition
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "msft_partition") {
			wmiResult, err = wmi.GetMSFTPartition("")
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetMSFTiSCSITargetPortal
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "msft_iscsitargetportal") {
			wmiResult, err = wmi.GetMSFTiSCSITargetPortal("")
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetMSIscsiInitiatorTargetClass
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "msiscsiinitiator_targetclass") {
			wmiResult, err = wmi.GetMSIscsiInitiatorTargetClass("")
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetMSiSCSIPortalInfoClass
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "msiscsi_portalinfoclass") {
			wmiResult, err = wmi.GetMSiSCSIPortalInfoClass()
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetWin32DiskDrive
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "win32_diskdrive") {
			wmiResult, err = wmi.GetWin32DiskDrive("")
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetWin32DiskPartition
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "win32_diskpartition") {
			wmiResult, err = wmi.GetWin32DiskPartition("")
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetWin32OperatingSystem
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "win32_operatingsystem") {
			wmiResult, err = wmi.GetWin32OperatingSystem()
			wmiResults = append(wmiResults, wmiResult)
		}

		// wmi.GetWin32Volume
		if wmiClass == queryAll || strings.EqualFold(wmiClass, "win32_volume") {
			wmiResult, err = wmi.GetWin32Volume()
			wmiResults = append(wmiResults, wmiResult)
		}

		// If we're querying all WMI classes, ignore any error
		if wmiClass == queryAll {
			err = nil
		}

		// If we have a non-zero loop count value, increment our index (0 means infinite loop)
		if loopCount != 0 {
			i++
		}
	}

	// Dump last enumeration to our log and console
	if (err == nil) && (len(wmiResults) > 0) {
		for _, wmiResult := range wmiResults {
			// Skip over any empty results
			if wmiResult == nil {
				continue
			}
			// Convert Go object to JSON
			var wmiTextResults []byte
			wmiTextResults, err = json.MarshalIndent(wmiResult, "", "    ")
			if err != nil {
				break
			}
			// Convert JSON to text and print to log file and console
			wmiText := string(wmiTextResults)
			log.Info("\n" + wmiText)
			fmt.Println(wmiText)
		}
	}

	// Log error and notify user if an error occurred
	if err != nil {
		sError := fmt.Sprintf("WMI processing failure, %v", err)
		log.Error(sError)
		fmt.Println(sError)
	}

	return err
}
