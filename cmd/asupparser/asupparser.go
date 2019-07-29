// (c) Copyright 2018 Hewlett Packard Enterprise Development LP
//
// This internal only use tool manually parses the ASUP folders looking for
// recorded host information phase 1 data.  Nimble Storage arrays are
// ignored.  Data collected is stored into a CSV file.

package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/hpe-storage/common-host-libs/asupparser"
	log "github.com/hpe-storage/common-host-libs/logger"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// Enumerate files in a given folder
///////////////////////////////////////////////////////////////////////////////

// enumerateFiles returns list of files under give folder, ignoring sub directories
func enumerateFiles(rootFolder string) []string {

	// Log entry/exit of routine
	log.Tracef("EnumerateFiles Enter, rootFolder=%v", rootFolder)

	var filePaths []string

	files, err := ioutil.ReadDir(rootFolder)
	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			filePath := filepath.Join(rootFolder, file.Name())
			filePaths = append(filePaths, filePath)
		}
	}

	// Log entry/exit of routine
	log.Tracef("EnumerateFiles Exit, count=%v", len(filePaths))

	return filePaths
}

// copyFile copies the give source file to destination
func copyFile(src, dst string) error {

	// Log entry/exit of routine
	log.Tracef("CopyFile Enter, src=%v, dst=%v", src, dst)

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// Log entry/exit of routine
	log.Traceln("CopyFile Exit")

	return out.Close()
}

// decompressFile decompresses given source file into destination folder using gunzip or tar based on src extension
func decompressFile(srcFile string, destFolder string) (err error) {
	// Log entry/exit of routine
	log.Tracef("DecompressFile Enter, srcFile=%v, destFolder=%v", srcFile, destFolder)

	var cmd *exec.Cmd
	if !strings.HasSuffix(srcFile, "tgz") && strings.HasSuffix(srcFile, ".gz") {
		args := []string{srcFile}
		cmd = exec.Command("gunzip", args...)
	} else {
		args := []string{"-xvzf", srcFile, "-C", destFolder, "--strip-components", "5"}
		cmd = exec.Command("tar", args...)
	}
	yo, err2 := cmd.CombinedOutput()
	if err2 != nil {
		fmt.Println(string(yo))
		fmt.Println(err2)
		return errors.New("DecompressFile Failed, err :" + err2.Error())
	}

	// Log entry/exit of routine
	log.Traceln("DecompressFile Exit")
	return nil
}

// extractHostInformation extracts host information from the raw data log file
func extractHostInformation(rawData string) []asupparser.XMLDataEntry {
	// Log entry/exit of routine
	log.Traceln("ExtractHostInformation Enter")

	var hiInfo []asupparser.XMLDataEntry

	entries := strings.SplitAfter(rawData, "</root>")
	for _, entry := range entries {

		entryBytes := []byte(entry)
		entryBytes = bytes.Trim(entryBytes, "\x00")
		entry := string(entryBytes[:])

		if len(strings.TrimSpace(entry)) == 0 {
			break
		}

		data := strings.SplitAfter(entry, "</itn_info>")
		if len(data) != 2 {
			panic("Unexpected host information!")
		}

		arrayHeader := strings.TrimSpace(data[0])
		hostInfo := strings.TrimSpace(data[1])

		// Convert the input XML text into an data into a XmlNodeHostInfo structure
		v := asupparser.XMLArrayHeader{}
		byteData := []byte(arrayHeader)
		err := xml.Unmarshal(byteData, &v)
		if err != nil {
			fmt.Println(arrayHeader)
			fmt.Println(err)
			panic(err)
		}

		if len(v.TimeStamp) != 24 {
			panic("Invalid XML timestamp length")
		}

		var xmlDataEntry asupparser.XMLDataEntry
		xmlDataEntry.TimeStamp = v.TimeStamp
		xmlDataEntry.XMLHostInfo = hostInfo

		hiInfo = append(hiInfo, xmlDataEntry)
	}

	// Log entry/exit of routine
	log.Tracef("ExtractHostInformation Exit, count=%v", len(hiInfo))

	return hiInfo
}

// parseLinuxHostInformation parses the XML input into a Linux HostInformation structure
func parseLinuxHostInformation(xmlData string) (asupparser.HostInfoRoot, error) {

	// Log entry/exit of routine
	log.Traceln("ParseLinuxHostInformation Enter")

	var hiInfo asupparser.HostInfoRoot
	byteData := []byte(xmlData)
	err := xml.Unmarshal(byteData, &hiInfo)
	if err != nil {
		fmt.Println("Unable to unmarshall linux host info " + err.Error())
		return hiInfo, err
	}

	// Log entry/exit of routine
	log.Traceln("ParseLinuxHostInformation Exit")
	return hiInfo, nil
}

// parseWindowsHostInformation parses the XML input into a Windows HostInformation structure
func parseWindowsHostInformation(xmlData string) (asupparser.HostInformation, error) {

	// Log entry/exit of routine
	log.Traceln("ParseWindowsHostInformation Enter")

	var hostInformation asupparser.HostInformation
	var jsonByteArray []byte

	// Convert the input XML text into an data into a XmlNodeHostInfo structure
	v := asupparser.XMLNodeHostInfo{}
	err := xml.Unmarshal([]byte(xmlData), &v)
	if err != nil {
		log.Errorf("Unable to parse XmlNodeHostInfo, err=%v", err)
		return hostInformation, err
	}

	// Extract the SystemInfo data
	jsonByteArray = []byte(strings.TrimSpace(v.SystemInfo))
	err = json.Unmarshal(jsonByteArray, &hostInformation.SystemInfo)
	if err != nil {
		log.Errorf("Unable to parse hostInformation.SystemInfo, err=%v", err)
	}

	// Extract the OS data
	jsonByteArray = []byte(strings.TrimSpace(v.OS))
	err = json.Unmarshal(jsonByteArray, &hostInformation.OS)
	if err != nil {
		log.Errorf("Unable to parse hostInformation.OS, err=%v", err)
	}

	// Extract the Windows feature data
	jsonByteArray = []byte(strings.TrimSpace(v.GetWindowsFeature))
	err = json.Unmarshal(jsonByteArray, &hostInformation.GetWindowsFeature)
	if err != nil {
		log.Errorf("Unable to parse hostInformation.GetWindowsFeature, err=%v", err)
	}

	// Extract the MPIO registered DSM data
	jsonByteArray = []byte(strings.TrimSpace(v.MpioRegisteredDsm))
	err = json.Unmarshal(jsonByteArray, &hostInformation.MpioRegisteredDsms)
	if err != nil {
		log.Errorf("Unable to parse hostInformation.MpioRegisteredDsms, err=%v", err)
	}

	// Log entry/exit of routine
	log.Tracef("ParseWindowsHostInformation Exit, err=%v", err)

	return hostInformation, err
}

// linuxHostInformationToMap Convert the given Linux HostInformation structure into a string map
func linuxHostInformationToMap(hostInformation asupparser.HostInfoRoot) map[string]string {

	// Log entry/exit of routine
	log.Traceln("LinuxHostInformationToMap Enter")

	mapData := make(map[string]string)
	mapData["SystemInfoName"] = strings.Replace(hostInformation.SystemInfo.Hostname.Text, ",", "_", -1)
	mapData["SystemInfoManufacturer"] = strings.Replace(hostInformation.SystemInfo.Manufacturer.Text, ",", "_", -1)
	mapData["SystemInfoModel"] = strings.Replace(hostInformation.SystemInfo.ProductName.Text, ",", "_", -1)
	mapData["SystemOsName"] = strings.Replace(hostInformation.OS.Distro.Text, ",", "_", -1)
	mapData["SystemOsVersion"] = strings.Replace(hostInformation.OS.Version.Text, ",", "_", -1)
	mapData["SystemKernelVersion"] = strings.Replace(hostInformation.OS.Kernel.Text, ",", "_", -1)

	mapData["NCM"] = strings.Replace(hostInformation.NLT.NCM.Text, ",", "_", -1)
	mapData["Scaleout"] = strings.Replace(hostInformation.NLT.NCMScaleout.Text, ",", "_", -1)
	mapData["Oracle"] = strings.Replace(hostInformation.NLT.Oracle.Text, ",", "_", -1)
	mapData["Docker"] = strings.Replace(hostInformation.NLT.Docker.Text, ",", "_", -1)
	mapData["NltVersion"] = strings.Replace(hostInformation.NLT.Attrversion, ",", "_", -1)
	mapData["MultipathVersion"] = strings.Replace(hostInformation.Multipath.Attrversion, ",", "_", -1)

	// Log entry/exit of routine
	log.Tracef("LinuxHostInformationToMap Exit, count=%v", len(mapData))

	return mapData
}

// linuxMultipathInformationToMap converts the given Linux HostInformation structure into a string map of multipath data
func linuxMultipathInformationToMap(hostInformation asupparser.HostInfoRoot) map[string]map[string]string {

	// Log entry/exit of routine
	log.Traceln("LinuxMultipathInformationToMap Enter")

	mapData := make(map[string]map[string]string)
	defaultsMap := make(map[string]string)
	for _, element := range hostInformation.Multipath.MultipathConf.MultipathDefaults.MultipathProperties.MultipathProperty {
		defaultsMap[element.Attrname] = strings.Trim(element.Attrvalue, "")
	}

	// add device section within blacklist section
	blacklistMap := make(map[string]string)
	for _, element := range hostInformation.Multipath.MultipathConf.MultipathBlackList.MultipathDevice.MultipathProperties.MultipathProperty {
		blacklistMap[element.Attrname] = strings.Trim(element.Attrvalue, "")
	}

	// add individual entires in blacklist section
	for _, element := range hostInformation.Multipath.MultipathConf.MultipathBlackList.MultipathEntries.MultipathProperties.MultipathProperty {
		blacklistMap[element.Attrname] = strings.Trim(element.Attrvalue, "")
	}

	// add device section within blacklist exceptions section
	blacklistExceptionMap := make(map[string]string)
	for _, element := range hostInformation.Multipath.MultipathConf.MultipathBlacklistExceptions.MultipathDevice.MultipathProperties.MultipathProperty {
		blacklistExceptionMap[element.Attrname] = strings.Trim(element.Attrvalue, "")
	}
	// add individual entires in blacklist exceptions section
	for _, element := range hostInformation.Multipath.MultipathConf.MultipathBlackList.MultipathDevice.MultipathProperties.MultipathProperty {
		blacklistExceptionMap[element.Attrname] = strings.Trim(element.Attrvalue, "")
	}

	deviceMap := make(map[string]string)
	for _, element := range hostInformation.Multipath.MultipathConf.MultipathDevices.MultipathDevice.MultipathProperties.MultipathProperty {
		deviceMap[element.Attrname] = strings.Trim(element.Attrvalue, "")
	}

	// populate each section now.
	mapData["defaults"] = defaultsMap
	mapData["blacklist"] = blacklistMap
	mapData["blacklistExceptions"] = blacklistExceptionMap
	mapData["devices"] = deviceMap

	// Log entry/exit of routine
	log.Tracef("LinuxMultipathInformationToMap Exit, count=%v", len(mapData))

	return mapData
}

// windowsHostInformationToMap convert the given HostInformation structure into a string map
func windowsHostInformationToMap(hostInformation asupparser.HostInformation) map[string]string {

	// Log entry/exit of routine
	log.Traceln("WindowsHostInformationToMap Enter")

	mapData := make(map[string]string)
	mapData["SystemInfoName"] = strings.Replace(hostInformation.SystemInfo.Name, ",", "_", -1)
	mapData["SystemInfoManufacturer"] = strings.Replace(hostInformation.SystemInfo.Manufacturer, ",", "_", -1)
	mapData["SystemInfoModel"] = strings.Replace(hostInformation.SystemInfo.Model, ",", "_", -1)
	mapData["SystemOsName"] = strings.Replace(hostInformation.OS.Name, ",", "_", -1)
	mapData["SystemOsVersion"] = strings.Replace(hostInformation.OS.Version, ",", "_", -1)

	for _, element := range hostInformation.GetWindowsFeature {
		keyName := "Windows" + strings.Replace(element.Name, " ", "", -1)
		if element.Installed {
			mapData[keyName] = "X"
		}
	}

	for _, element := range hostInformation.MpioRegisteredDsms.DsmParameters {
		keyName := "Windows" + strings.Replace(element.DsmName, " ", "", -1)
		mapData[keyName] = element.DsmVersion
	}

	// Log entry/exit of routine
	log.Tracef("WindowsHostInformationToMap Exit, count=%v", len(mapData))

	return mapData
}

// getHostInformation returns parsed information for a given asup array folder
func getHostInformation(asupFolder string) (arrayHostInfo map[string]string, arrayMultipathInfo map[string]string, err error) {
	arrayHostInfo = make(map[string]string)
	arrayMultipathInfo = make(map[string]string)
	// Log the ASUP folder details
	log.Infof("asupFolder=%v", asupFolder)

	// Get the path to the array logs
	pathArrayLogTgz := filepath.Join(asupFolder, "/array_log.tgz")
	// Get the path to the array logs
	log.Infof("pathArrayLogTgz=%v", pathArrayLogTgz)

	if _, err = os.Stat(pathArrayLogTgz); os.IsNotExist(err) {
		log.Errorf("Skipping array_log.tgz, err=%v", err)
		return nil, nil, err
	}

	// Initialize our temp folder
	log.Tracef("tempFolder=%v", *tmpFolder)
	os.RemoveAll(*tmpFolder)
	err = os.MkdirAll(*tmpFolder, 0644)
	if err != nil {
		fmt.Printf("Error creating temp folder, err=%v", err)
		panic(err)
	}

	// For performance gains, copy the ASUP array_log.tgz to temp folder
	localArrayLogTgz := filepath.Join(*tmpFolder, "array_log.tgz")
	copyFile(pathArrayLogTgz, localArrayLogTgz)

	// Decompress array_log.tgz into our temp folder
	decompressFile(localArrayLogTgz, *tmpFolder)

	// Enumerate all the array log files
	arrayLogFiles := enumerateFiles(*tmpFolder)
	for _, arrayLogFile := range arrayLogFiles {

		// Decompress any compressed hi_info_collect log file
		_, fileName := filepath.Split(arrayLogFile)
		if strings.HasPrefix(fileName, "hi_info_collect.log.") {
			log.Infof("Decompressing %v", arrayLogFile)
			decompressFile(arrayLogFile, *tmpFolder)
		}
	}

	// Re-enumerate all the array log files
	arrayLogFiles = enumerateFiles(*tmpFolder)
	for _, arrayLogFile := range arrayLogFiles {

		_, fileName := filepath.Split(arrayLogFile)
		// If log file doesn't start with "hi_info_collect.log", ignore it
		if !strings.HasPrefix(fileName, "hi_info_collect.log") {
			continue
		}

		// If it's our compressed log file, ignore it
		if strings.HasSuffix(fileName, ".gz") {
			continue
		}

		// Read in our host information log file
		log.Tracef("Read hi_info_collect into memory, arrayLogFile=%v", arrayLogFile)
		fileData, err := ioutil.ReadFile(arrayLogFile)
		if err != nil {
			panic(err)
		}

		// Extract the XML data from the host information file
		xmlHostInformation := extractHostInformation(string(fileData))

		// Parse through each XML entry
		for _, hostInfo := range xmlHostInformation {
			// ignore any errors and continue with other hosts
			populateHostInformation(hostInfo, arrayHostInfo, arrayMultipathInfo)
		}
	}
	return arrayHostInfo, arrayMultipathInfo, nil
}

// populateHostInformation populates host info and multipath info maps from given host info data
func populateHostInformation(hostInfo asupparser.XMLDataEntry, arrayHostInfo map[string]string, arrayMultipathInfo map[string]string) (err error) {
	if strings.Contains(hostInfo.XMLHostInfo, "Windows") {
		// Parse XML entry into HostInformation structure
		hostInformation, err := parseWindowsHostInformation(hostInfo.XMLHostInfo)
		if err != nil {
			log.Errorln("Unable to parse windows host info, err: ", err.Error())
			return err
		}
		// Convert HostInformation structure into string map
		m := windowsHostInformationToMap(hostInformation)
		// Form a unique key for our host to avoid duplicate entries
		keyUnique := fmt.Sprintf("%v%v%v%v", m["SystemInfoName"], m["SystemInfoManufacturer"], m["SystemInfoModel"], m["SystemOsVersion"])
		keyUnique = strings.Replace(keyUnique, " ", "", -1)
		log.Tracef("keyUnique=%v", keyUnique)

		// Add the CSV entry to our host information map
		arrayHostInfo[keyUnique] = fmt.Sprintf("%v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v", "Host", true, m["SystemOsName"], m["SystemOsVersion"], m["SystemInfoName"], m["SystemInfoManufacturer"], m["SystemInfoModel"], m["WindowsHyper-V"], m["WindowsMultipath-IO"], m["WindowsMicrosoftDSM"], m["WindowsNimbleDSM"])
	} else {
		// Parse XML entry into HostInformation structure
		hostInformation, err := parseLinuxHostInformation(hostInfo.XMLHostInfo)
		if err != nil {
			log.Errorln("Unable to parse linux host info, err: ", err.Error())
			return err
		}
		// Convert HostInformation structure into string map
		m := linuxHostInformationToMap(hostInformation)
		// Form a unique key for our host to avoid duplicate entries
		keyUnique := fmt.Sprintf("%v%v%v%v", m["SystemInfoName"], m["SystemInfoManufacturer"], m["SystemInfoModel"], m["SystemOsVersion"])
		keyUnique = strings.Replace(keyUnique, " ", "", -1)
		log.Tracef("keyUnique=%v", keyUnique)

		// Add the CSV entry to our host information map
		arrayHostInfo[keyUnique] = fmt.Sprintf("%v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v %v, %v, %v, %v, %v, %v, %v, %v", "Host", true, m["SystemOsName"], m["SystemOsVersion"], m["SystemInfoName"], m["SystemInfoManufacturer"], m["SystemInfoModel"], "", "", "", "", "", m["SystemKernelVersion"], m["NCM"], m["Scaleout"], m["Oracle"], m["Docker"], m["NltVersion"], m["MultipathVersion"])

		// Create an entry for multipath information
		multipathMap := linuxMultipathInformationToMap(hostInformation)
		for section, mapData := range multipathMap {
			for name, value := range mapData {
				keyUnique = keyUnique + section + name
				rowType := "multipath"
				arrayMultipathInfo[keyUnique] = fmt.Sprintf("%v, %v, %v, %v, %v", m["SystemInfoName"], rowType, section, name, value)
			}
		}
	}
	return nil
}

// createCSV checks and creates required CSV file
func createCSV(filename string) (fileCSV *os.File, err error) {
	// Create the output CSV folder if it isn't already present
	csvFolder, _ := filepath.Split(filename)
	if _, err = os.Stat(csvFolder); os.IsNotExist(err) {
		err = os.MkdirAll(csvFolder, 0644)
		if err != nil {
			panic(err)
		}
		log.Infof("CSV created, csvFolder=%v", csvFolder)
	} else {
		log.Infof("CSV folder already present, csvFolder=%v", csvFolder)
	}

	// Open output CSV and write out header
	fileCSV, err = os.Create(filename)
	if err != nil {
		log.Errorf("unable to create CSV file %s", filename)
		return nil, err
	}
	return fileCSV, nil
}

// initialize command options for short options
func init() {
	flag.StringVar(outCSV, "out", "/auto/share/asupparser/hi_phase1.csv", "File path where output CSV will be stored.")
	flag.StringVar(multipathCSV, "multipath", "/auto/share/asupparser/hi_phase1_multipath.csv", "File path where multipath output CSV will be stored.")
	flag.StringVar(rootFolder, "asup", "", "Root folder of ASUP data.")
	flag.StringVar(logFile, "log", "/auto/share/asupparser/hi_phase1_parser.log", "log file")
	flag.StringVar(tmpFolder, "temp", "/tmp/asupparser", "Temp folder to extract array logs")
}

var (
	// Configure and parse the CLI input
	outCSV       = flag.String("output-csv", "/auto/share/asupparser/hi_phase1.csv", outputCSVDescription)
	rootFolder   = flag.String("asup-folder", "/auto/support/autosupport/san", asupFolderDescription)
	logFile      = flag.String("log-file", "/auto/share/asupparser/hi_phase1_parser.log", "log file")
	multipathCSV = flag.String("multipath-csv", "/auto/share/asupparser/hi_phase1_multipath.csv", multipathCSVDescription)
	tmpFolder    = flag.String("temp-folder", "/tmp/asupparser", tempDescription)
)

const (
	asupFolderDescription   = "Root folder of ASUP data."
	outputCSVDescription    = "File path where output CSV will be stored. Default: /auto/share/asupparser/hi_phase1.csv"
	multipathCSVDescription = "File path where linux multipath output CSV will be stored. Default: /auto/share/asupparser/hi_phase1_multipath.csv"
	logFileDescription      = "File path where debug log will be stored. Default: /auto/share/asupparser/hi_phase1_parser.log"
	tempDescription         = "Temp folder to extract array logs. Default: /tmp/asupparser"
)

// Program entry point
func main() {

	// override Usage
	flag.Usage = func() {
		fmt.Printf("\nHost Information Phase 1 Data Collector\n")
		fmt.Printf("\nUsage:\n")
		fmt.Println()
		fmt.Printf("asupparser [--asup-folder|--asup <array folder>] [--output-csv|--out <filename>] [--multipath-csv|--multipath <filename>] [--log-file|--log <filename>] [--temp-folder|--temp <folder>]\n")
		fmt.Printf("\nOptions:\n")
		fmt.Printf("\t%-30s\t%-50s\n", "-asup, -asup-folder", asupFolderDescription)
		fmt.Printf("\t%-30s\t%-50s\n", "-out, -output-csv", outputCSVDescription)
		fmt.Printf("\t%-30s\t%-50s\n", "-multipath, -multipath-csv", multipathCSVDescription)
		fmt.Printf("\t%-30s\t%-50s\n", "-log, -log-file", logFileDescription)
		fmt.Printf("\t%-30s\t%-50s\n", "-temp, -temp-folder", tempDescription)
		fmt.Println()
	}

	// parse cli options
	flag.Parse()

	if flag.Parsed() != true {
		fmt.Println("Error parsing command options")
		flag.Usage()
		os.Exit(2)
	}

	log.InitLogging(*logFile, &log.LogParams{Level: "trace"}, false)

	// Log the input parameters
	log.Tracef("CLI input - outCSV       = %v", *outCSV)
	log.Tracef("CLI input - rootFolder   = %v", *rootFolder)
	log.Tracef("CLI input - logFile   = %v", *logFile)
	log.Tracef("CLI input - multipathCSV   = %v", *multipathCSV)
	log.Tracef("CLI input - tmpFolder   = %v", *tmpFolder)

	// Insert a separator line between invocations
	fmt.Printf("Parsing host information from %s...\n", *rootFolder)

	var mCSV *os.File
	var fCSV *os.File
	var err error

	fCSV, err = createCSV(*outCSV)
	if err != nil {
		return
	}
	defer fCSV.Close()
	fmt.Fprintln(fCSV, "RowType, HaveHostInfo, SystemOsName, SystemOsVersion, SystemInfoName, SystemInfoManufacturer, SystemInfoModel, WindowsHyper-V, WindowsMultipath-IO, WindowsMicrosoftDSM, WindowsNimbleDSM, SystemKernelVersion, NCM, Scaleout, Oracle, Docker, NltVersion, MultipathVersion")

	mCSV, err = createCSV(*multipathCSV)
	if err != nil {
		return
	}
	defer mCSV.Close()
	fmt.Fprintln(mCSV, "SystemOsName, RowType, SectionName, PropertyName, PropertyValue")

	arrayHostInfo, multipathInfo, err := getHostInformation(*rootFolder)
	if err != nil {
		log.Errorln("Unable to fetch host information for rootFolder " + *rootFolder + " err: " + err.Error())
		os.Exit(1)
	}
	// Now add all the enumerated host information to the array.
	for _, v := range arrayHostInfo {
		fmt.Fprintln(fCSV, v)
	}
	// add all hosts multipath information
	for _, v := range multipathInfo {
		fmt.Fprintln(mCSV, v)
	}

	fmt.Printf("Successfully completed parsing host configuration data, result copied as %s and %s\n", *outCSV, *multipathCSV)
	return
}
