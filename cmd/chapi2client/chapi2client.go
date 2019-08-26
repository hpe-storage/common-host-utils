// (c) Copyright 2019 Hewlett Packard Enterprise Development LP

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hpe-storage/common-host-libs/chapi2/model"
	log "github.com/hpe-storage/common-host-libs/logger"
)

const (
	// Application main menu (shows CHAPI2 endpoints)
	mainMenu = `
    1.)   GET    /api/v1/hosts
    2.)   GET    /api/v1/networks
    3.)   GET    /api/v1/initiators
    4.)   GET    /api/v1/devices
    5.)   GET    /api/v1/devices/details
    6.)   GET    /api/v1/devices/{serialNumber}/partitions
    7.)   POST   /api/v1/devices
    8.)   DELETE /api/v1/devices/{serialnumber}
    9.)   PUT    /api/v1/devices/{serialnumber}/actions/offline
    10.)  PUT    /api/v1/devices/{serialNumber}/{filesystem}
    11.)  GET    /api/v1/mounts
    12.)  GET    /api/v1/mounts/details
    13.)  POST   /api/v1/mounts
    14.)  DELETE /api/v1/mounts/{mountId}
`

	// Input options
	inputAccessProtocol = "Enter access protocol (fc, iscsi):  "
	inputEnterOption    = "Enter option:  "
	inputFileSystem     = "Enter file system:  "
	inputMountPoint     = "Enter mount point:"
	inputMountPointID   = "Enter mount point ID:  "
	inputSerialNumber   = "Enter device serial number:  "

	// iSCSI input options
	inputIscsiChapPassword = "Enter CHAP Password:  "
	inputIscsiChapUser     = "Enter CHAP User:  "
	inputIscsiConnectType  = "Enter iSCSI connect type (ping, subnet, auto_initiator, or leave empty):  "
	inputIscsiDiscoveryIP  = "Enter Discovery IP:  "
	inputIscsiTargetName   = "Enter iSCSI target name:  "
	inputIscsiTargetScope  = "Enter iSCSI target scope (volume, group):  "
)

func main() {
	// Create log file
	chapiLogFile := filepath.Join(getLogPath(), "chapi2client.log")
	log.InitLogging(chapiLogFile, &log.LogParams{Level: "trace"}, false)

	// Create a CHAPI client object
	chapiClient, err := getChapiClient()
	if err != nil {
		return
	}

	for {
		// Display main menu and prompt for option
		var inputString string
		fmt.Print(mainMenu + "\n")
		if inputString, err = readString(inputEnterOption); err != nil {
			fmt.Println(err)
			continue
		}

		// Ignore blank line input
		if inputString == "" {
			continue
		}

		// Convert user input to a number.  If any non-integer value is entered, exit program.
		var inputOption int
		if inputOption, err = strconv.Atoi(inputString); err != nil {
			return
		}

		// Handle each CHAPI2 endpoint request
		switch inputOption {
		case 1:
			dumpChapiObject(chapiClient.GetHostInfo())
		case 2:
			dumpChapiObject(chapiClient.GetHostNetworks())
		case 3:
			dumpChapiObject(chapiClient.GetHostInitiators())
		case 4:
			dumpChapiObject(chapiClient.GetDevices(""))
		case 5:
			if serialNumber, err := readString(inputSerialNumber); err == nil {
				dumpChapiObject(chapiClient.GetAllDeviceDetails(serialNumber))
			}
		case 6:
			if serialNumber, err := readString(inputSerialNumber); (err == nil) && (serialNumber != "") {
				dumpChapiObject(chapiClient.GetPartitionInfo(serialNumber))
			}
		case 7:
			if publishInfo := getPublishObject(); publishInfo != nil {
				dumpChapiObject(chapiClient.CreateDevice(*publishInfo))
			}
		case 8:
			if serialNumber, err := readString(inputSerialNumber); (err == nil) && (serialNumber != "") {
				dumpChapiObject(nil, chapiClient.DeleteDevice(serialNumber))
			}
		case 9:
			if serialNumber, err := readString(inputSerialNumber); (err == nil) && (serialNumber != "") {
				dumpChapiObject(nil, chapiClient.OfflineDevice(serialNumber))
			}
		case 10:
			if serialNumber, err := readString(inputSerialNumber); (err == nil) && (serialNumber != "") {
				if fileSystem, err := readString(inputFileSystem); (err == nil) && (fileSystem != "") {
					dumpChapiObject(nil, chapiClient.CreateFileSystem(serialNumber, fileSystem))
				}
			}
		case 11:
			dumpChapiObject(chapiClient.GetMounts(""))
		case 12:
			var serialNumber, mountPointID string
			if serialNumber, err = readString(inputSerialNumber); err == nil {
				if serialNumber != "" {
					mountPointID, _ = readString(inputMountPointID)
				}
				dumpChapiObject(chapiClient.GetAllMountDetails(serialNumber, mountPointID))
			}
		case 13:
			if serialNumber, err := readString(inputSerialNumber); (err == nil) && (serialNumber != "") {
				if mountPoint, err := readString(inputMountPoint); (err == nil) && (mountPoint != "") {
					dumpChapiObject(chapiClient.CreateMount(serialNumber, mountPoint, nil))
				}
			}
		case 14:
			if serialNumber, err := readString(inputSerialNumber); (err == nil) && (serialNumber != "") {
				if mountPointID, err := readString(inputMountPointID); (err == nil) && (mountPointID != "") {
					dumpChapiObject(nil, chapiClient.DeleteMount(serialNumber, mountPointID))
				}
			}
		}
	}
}

// getPublishObject is used to inialize a model.PublishInfo to create a new block device
func getPublishObject() *model.PublishInfo {
	var serialNumber, accessProtocol string
	var err error

	// Get serial number (mandatory field)
	if serialNumber, err = readString(inputSerialNumber); (err != nil) || (serialNumber == "") {
		return nil
	}

	// Get access protocol (mandatory field)
	if accessProtocol, err = readString(inputAccessProtocol); (err != nil) || ((accessProtocol != model.AccessProtocolFC) && (accessProtocol != model.AccessProtocolIscsi)) {
		return nil
	}

	// Allocate and initialize a base model.PublishInfo object
	publishInfo := &model.PublishInfo{
		SerialNumber: serialNumber,
		BlockDev: &model.BlockDeviceAccessInfo{
			AccessProtocol: accessProtocol,
		},
	}

	// If this is not an iSCSI device (e.g. FC), object is fully initialized
	if accessProtocol != model.AccessProtocolIscsi {
		return publishInfo
	}

	// It's an iSCSI device so query the target IQN and target scope
	publishInfo.BlockDev.TargetName, _ = readString(inputIscsiTargetName)
	publishInfo.BlockDev.TargetScope, _ = readString(inputIscsiTargetScope)

	// Get the model.IscsiAccessInfo object values
	var connectType, discoveryIP, chapUser, chapPassword string
	connectType, _ = readString(inputIscsiConnectType)
	discoveryIP, _ = readString(inputIscsiDiscoveryIP)
	if chapUser, _ = readString(inputIscsiChapUser); chapUser != "" {
		chapPassword, _ = readString(inputIscsiChapPassword)
	}

	// Allocate and add the model.IscsiAccessInfo object to the BlockDev object
	publishInfo.BlockDev.IscsiAccessInfo = &model.IscsiAccessInfo{
		ConnectType:  connectType,
		DiscoveryIP:  discoveryIP,
		ChapUser:     chapUser,
		ChapPassword: chapPassword,
	}

	// Return the iSCSI initialized model.PublishInfo object
	return publishInfo
}

// readString is used to request an input string from the user
func readString(prompt string) (line string, err error) {

	// Prompt the user (if prompt provided)
	reader := bufio.NewReader(os.Stdin)
	if prompt != "" {
		fmt.Print(prompt)
	}

	// Read the string the user entered
	if line, err = reader.ReadString('\n'); err != nil {
		return "", nil
	}

	// Strip CRLF for Windows to deal with the string properly
	line = strings.Replace(line, "\r\n", "", -1)

	// Return the user entered line
	return line, nil
}

// dumpChapiObject takes the CHAPI return data, and error.  If an error occurred, the error is
// displayed.  If no error, and a CHAPI object was returned, the CHAPI object is converted to
// JSON before being displayed.
func dumpChapiObject(data interface{}, err error) {

	// If an error occurred, display the error and return
	if err != nil {
		fmt.Println(err)
		return
	}

	// If a CHAPI data object was returned, convert to JSON and display
	if data != nil {
		b, errJSON := json.MarshalIndent(data, "", "    ")
		if errJSON != nil {
			fmt.Println(errJSON)
			return
		}
		fmt.Println(string(b))
	}
}
