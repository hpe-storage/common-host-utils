// (c) Copyright 2019 Hewlett Packard Enterprise Development LP

// Windows Go binaries can have version information attached to them.  A Go package is available
// to help you attach this information:
//
// https://github.com/josephspurrier/goversioninfo
//
// What this utility does is take an existing versioninfo.json and does an in-place update of
// that file's applicable properties.  The following properties are adjusted:
//
//    * Company name
//    * Legal Copyright
//    * File version
//    * Product version
//
// CLI options provide this appilcation the version details.  For example:
//
// updatewinversioninfo --version=v1.2.3 --build=456 ./versioninfo.json
//
// In the above example, the Windows file/product version (in versioninfo.json) will be set to 1.2.3.456

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/josephspurrier/goversioninfo"
)

// Subset of goversioninfo.VersionInfo.  These are the only properties that are utilized
// in our vesrioninfo.json file.
type versionInfo struct {
	goversioninfo.FixedFileInfo  `json:"FixedFileInfo"`
	goversioninfo.StringFileInfo `json:"StringFileInfo"`
	goversioninfo.VarFileInfo    `json:"VarFileInfo"`
	IconPath                     string `json:"IconPath"`
	ManifestPath                 string `json:"ManifestPath"`
}

const (
	defaultVersionInfoFilename = "versioninfo.json"
	envVersion                 = "VERSION"      // Environment variable to check for x.y.z version
	envBuild                   = "BUILD_NUMBER" // Environment variable to check for build number
)

func main() {

	var err error

	// Parse our supported version/build flags
	var version = flag.String("version", "", "help message for version")
	var build = flag.Int("build", -1, "help message for build")
	flag.Parse()

	// If --version override is not provided, try and retrieve the x.y.z version from the environment variable
	if (version == nil) || (*version == "") {
		*version = os.Getenv(envVersion)
		if *version == "" {
			fmt.Println("Version environment variable not set or --version override not provided")
			os.Exit(1)
		}
	}

	// If --build override is not provided, try and retrieve the build number frofrom the environment variable
	if (build == nil) || (*build == -1) {
		*build, err = strconv.Atoi(os.Getenv(envBuild))
		if (err != nil) || (*build < 0) {
			fmt.Println("Build environment variable not set or --build override not provided")
			os.Exit(1)
		}
	}

	// We fill in the Windows major, minor, patch, build version into this array
	var verInt [4]int

	// Split the version string into an array
	verString := strings.Split(strings.TrimLeft(*version, "v"), ".")
	if len(verString) != 3 {
		fmt.Printf("Invalid version string provided, version=%v\n", *version)
		os.Exit(1)
	}

	// Now convert version string into its numeric value (must fix within 16-bits)
	for i, v := range verString {
		verInt[i], err = strconv.Atoi(v)
		if err != nil {
			fmt.Printf("Unable to convert version string to int, err=%v\n", err)
			os.Exit(1)
		}
		if (verInt[i] < 0) || (verInt[i] > 0xFFFF) {
			fmt.Printf("Version value exceeds 16-bit limit, value=%v\n", verInt[i])
			os.Exit(1)
		}
	}

	// Set the build number
	if (*build < 0) || (*build > 0xFFFF) {
		fmt.Printf("Build number exceeds 16-bit limit, value=%v\n", build)
		os.Exit(1)
	}
	verInt[3] = *build

	// Retrieve the path to the input JSON file.  If none provided, we'll default to current
	// folder with a filename of versioninfo.json.
	versionFile := flag.Arg(0)
	if versionFile == "" {
		versionFile = defaultVersionInfoFilename
	}

	// Read in the entire JSON file
	var jsonData []byte
	jsonData, err = ioutil.ReadFile(versionFile)
	if err != nil {
		fmt.Printf("Unable to read %v, err=%v\n", versionFile, err)
		os.Exit(1)
	}

	// Unmarshal JSON data into our verInfo object
	verInfo := &versionInfo{}
	if err := json.Unmarshal([]byte(jsonData), &verInfo); err != nil {
		fmt.Printf("Unable to unmarshal JSON data, err=%v\n", err)
		os.Exit(1)
	}

	// Set/override the company name and legal copyright
	verInfo.StringFileInfo.CompanyName = "Hewlett Packard Enterprise"
	verInfo.StringFileInfo.LegalCopyright = fmt.Sprintf("(c) Copyright %v Hewlett Packard Enterprise Development LP", time.Now().Year())

	// Set the file/product version strings
	verInfo.StringFileInfo.FileVersion = fmt.Sprintf("%v.%v.%v.%v", verInt[0], verInt[1], verInt[2], verInt[3])
	verInfo.StringFileInfo.ProductVersion = verInfo.StringFileInfo.FileVersion

	// Set the file/product version strings
	verInfo.FixedFileInfo.FileVersion = goversioninfo.FileVersion{Major: verInt[0], Minor: verInt[1], Patch: verInt[2], Build: verInt[3]}
	verInfo.FixedFileInfo.ProductVersion = verInfo.FixedFileInfo.FileVersion

	// Marshal the object back into "pretty" JSON
	jsonData, err = json.MarshalIndent(verInfo, "", "    ")
	if err != nil {
		fmt.Printf("Could not marshal the JSON data, err=%v\n", err)
		os.Exit(1)
	}

	// Update the JSON file
	err = ioutil.WriteFile(versionFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Unable to write %v, err=%v\n", versionFile, err)
		os.Exit(1)
	}

	// On success we exit with 0, non-zero on any failure
	os.Exit(0)
}
