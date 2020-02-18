package main

// Copyright 2019 Hewlett Packard Enterprise Development LP.
import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/hpe-storage/common-host-libs/linux"
	log "github.com/hpe-storage/common-host-libs/logger"
	"github.com/hpe-storage/common-host-libs/tunelinux"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

var (
	recommendations []*tunelinux.Recommendation
)

const (
	// Red Red Colored Text
	Red = "\x1b[31;1m"
	// Green Green Colored Text
	Green = "\x1b[32;1m"
	// Yellow Yellow Colored Text
	Yellow = "\x1b[33;1m"
	// NoColor end colored text
	NoColor = "\x1b[0m"
	// XML format output
	XML = "xml"
	// JSON format output
	JSON = "json"
)

// Options for get sub command
type getOptions struct {
	category string
	status   string
	severity string
	verbose  bool
	json     bool
	xml      bool
}

// Options for set sub command
type setOptions struct {
	category string
	severity string
	global   bool
}

// change display color for text based on compliance status
func setColoredComplianceStatus(recommendation *tunelinux.Recommendation) {
	if recommendation.CompliantStatus != tunelinux.ComplianceStatus.String(tunelinux.Recommended) {
		if recommendation.Level == tunelinux.Severity.String(tunelinux.Info) {
			recommendation.CompliantStatus = fmt.Sprintf("%s%s%s", Yellow, recommendation.CompliantStatus, NoColor)
		} else if recommendation.Level == tunelinux.Severity.String(tunelinux.Warning) {
			recommendation.CompliantStatus = fmt.Sprintf("%s%s%s", Yellow, recommendation.CompliantStatus, NoColor)
		} else if recommendation.Level == tunelinux.Severity.String(tunelinux.Error) {
			recommendation.CompliantStatus = fmt.Sprintf("%s%s%s", Red, recommendation.CompliantStatus, NoColor)
		} else if recommendation.Level == tunelinux.Severity.String(tunelinux.Critical) {
			recommendation.CompliantStatus = fmt.Sprintf("%s%s%s", Red, recommendation.CompliantStatus, NoColor)
		}
	} else {
		recommendation.CompliantStatus = fmt.Sprintf("%s%s%s    ", Green, recommendation.CompliantStatus, NoColor)
	}
}

// printCategoryHeader will print header and separation between different category of recommendations
func printCategoryHeader(prevCategory, category string) {
	if prevCategory != category {
		if prevCategory != "" {
			// we alredy printed some categories, add separation between each category of recommendations
			fmt.Println("+------------+----------------------+----------------------+----------------------+----------------------+-----------------+-----------------+")
			fmt.Println()
		}
		fmt.Printf("Recommendations for %s:\n", category)
		fmt.Println("+------------+----------------------+----------------------+----------------------+----------------------+-----------------+-----------------+")
		fmt.Printf("| %-10s | %-20s | %-20s | %-20s | %-20s | %-15s | %-15s |\n", "Category", "Device", "Parameter", "Value", "Recommendation", "Status", "Severity")
		fmt.Println("+------------+----------------------+----------------------+----------------------+----------------------+-----------------+-----------------+")
	}
}

// displayTabularFormat : display recommendations in tabular format
func displayTabularFormat(complianceStatus string, severity string) {
	var category = ""
	if len(recommendations) > 0 {
		for _, recommendation := range recommendations {
			if recommendation != nil {
				if complianceStatus != tunelinux.All && recommendation.CompliantStatus != complianceStatus {
					continue
				}
				if severity != tunelinux.All && recommendation.Level != severity {
					continue
				}
				// check and print if new category of recommendations
				printCategoryHeader(category, recommendation.Category)
				if terminal.IsTerminal(int(os.Stdout.Fd())) {
					setColoredComplianceStatus(recommendation)
				}
				fmt.Printf("| %-10s | %-20s | %-20s | %-20s | %-20s | %-15s | %-15s |\n", recommendation.Category, recommendation.Device, recommendation.Parameter, recommendation.Value, recommendation.Recommendation, recommendation.CompliantStatus, recommendation.Level)
				category = recommendation.Category
			}
		}
		fmt.Println("+------------+----------------------+----------------------+----------------------+----------------------+-----------------+-----------------+")
	} else {
		fmt.Printf("No recommendations can be found. If running on a virtual machine with only VMDK/RDM devices, please check %s file to set disk queue settings manually\n", tunelinux.GetUdevTemplateFile())
	}
}

// displayVerboseFormat : display recommendations in verbose format
func displayVerboseFormat(complianceStatus string, severity string) {
	for _, recommendation := range recommendations {
		if recommendation != nil {
			if ignoreRecommendation(complianceStatus, severity, recommendation) == true {
				continue
			}
			if terminal.IsTerminal(int(os.Stdout.Fd())) {
				setColoredComplianceStatus(recommendation)
			}
			fmt.Printf("%-20s : %s\n", "Category", recommendation.Category)
			if recommendation.Category == tunelinux.Category.String(tunelinux.Filesystem) || recommendation.Category == tunelinux.Category.String(tunelinux.Disk) {
				fmt.Printf("%-20s : %s\n", "Device", recommendation.Device)
			}
			fmt.Printf("%-20s : %s\n", "Parameter", recommendation.Parameter)
			fmt.Printf("%-20s : %s\n", "Value", recommendation.Value)
			fmt.Printf("%-20s : %s\n", "Recommendation", recommendation.Recommendation)
			fmt.Printf("%-20s : %s\n", "ComplianceStatus", recommendation.CompliantStatus)
			fmt.Printf("%-20s : %s\n", "Severity", recommendation.Level)
			fmt.Printf("%-20s : %s\n", "Description", recommendation.Description)
			fmt.Println()
		}
	}
}

// ignore recommendations if the filters does not match
func ignoreRecommendation(complianceStatus string, severity string, recommendation *tunelinux.Recommendation) (ignore bool) {
	ignore = false
	if complianceStatus != tunelinux.All && recommendation.CompliantStatus != complianceStatus {
		ignore = true
	}
	if severity != tunelinux.All && recommendation.Level != severity {
		ignore = true
	}
	return ignore
}

// displayCustomFormat : display recommendations in either XML/JSON format
func displayCustomFormat(format string, complianceStatus string, severity string) {
	var finalRecommendations []*tunelinux.Recommendation
	var err error
	var result []byte
	for _, recommendation := range recommendations {
		if recommendation != nil {
			if ignoreRecommendation(complianceStatus, severity, recommendation) == true {
				continue
			}
			finalRecommendations = append(finalRecommendations, recommendation)
		}
	}
	if len(finalRecommendations) > 0 {
		if format == JSON {
			result, err = json.MarshalIndent(finalRecommendations, "", "\t")
		} else if format == XML {
			result, err = xml.MarshalIndent(finalRecommendations, "", "\t")
		}
		if err != nil {
			log.Errorf("Unable to convert recommendations to %s error: %s\n", format, err.Error())
			fmt.Printf("Error: Failed to convert recommendations to %s format, reason: %s\n", format, err.Error())
			return
		}
		fmt.Println(string(result))
	}
}

func getRecommendationByCategory(category string) (err error) {
	// Get All nimble devices
	devices, err := linux.GetNimbleDmDevices(false, "", "")
	if err != nil {
		log.Error("Unable to get Nimble devices ", err.Error())
		fmt.Print("Error: Unable to get Nimble devices ", err.Error())
		return err
	}
	switch category {
	case tunelinux.Category.String(tunelinux.Filesystem):
		recommendations, err = tunelinux.GetFileSystemRecommendations(devices)
	case tunelinux.Category.String(tunelinux.Multipath):
		recommendations, err = tunelinux.GetMultipathRecommendations()
	case tunelinux.Category.String(tunelinux.Disk):
		recommendations, err = tunelinux.GetDeviceRecommendations(devices)
	case tunelinux.Category.String(tunelinux.Iscsi):
		recommendations, err = tunelinux.GetIscsiRecommendations()
	case tunelinux.Category.String(tunelinux.Fc):
		recommendations, err = tunelinux.GetFcRecommendations()
	case tunelinux.All:
		recommendations, err = tunelinux.GetRecommendations()
	default:
		log.Error("Invalid category provided for recommendations")
		err = errors.New("Invalid category provided for recommendations")
	}
	if err != nil {
		log.Error("Unable to get recommendations for category ", category)
		err = errors.New("Error: Unable to get recommendations for category " + category + " error: " + err.Error())
	}
	return err
}

func setRecommendationsByCategory(category string, global bool) (err error) {
	if category == tunelinux.Category.String(tunelinux.Filesystem) || category == tunelinux.Category.String(tunelinux.Fc) {
		err = errors.New("Only multipath/disk/iscsi categories supported for set recommendations. For others please follow the documentation as specified in the description for each recommendation setting")
		return err
	}
	switch category {
	case tunelinux.Category.String(tunelinux.Multipath):
		err = tunelinux.SetMultipathRecommendations()
	case tunelinux.Category.String(tunelinux.Disk):
		err = tunelinux.SetBlockDeviceRecommendations()
	case tunelinux.Category.String(tunelinux.Iscsi):
		err = tunelinux.SetIscsiRecommendations(global)
	case tunelinux.All:
		err = tunelinux.SetRecommendations(global)
	}
	if err != nil {
		fmt.Printf("Failed to apply %s recommendations, error: %s\n", category, err.Error())
		return err
	}
	fmt.Printf("Successfully applied %s recommendations\n", category)
	return err
}

func validateCategory(category string) {
	var err error
	switch category {
	case tunelinux.All:
	case tunelinux.Category.String(tunelinux.Filesystem):
	case tunelinux.Category.String(tunelinux.Multipath):
	case tunelinux.Category.String(tunelinux.Disk):
	case tunelinux.Category.String(tunelinux.Fc):
	case tunelinux.Category.String(tunelinux.Iscsi):
	default:
		err = errors.New("Error: Invalid recommendation category provided")
	}
	if err != nil {
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

func validateStatus(complianceStatus string) {
	var err error
	switch complianceStatus {
	case tunelinux.All:
	case tunelinux.ComplianceStatus.String(tunelinux.Recommended):
	case tunelinux.ComplianceStatus.String(tunelinux.NotRecommended):
	default:
		err = errors.New("Error: Invalid compliance status provided")
	}
	if err != nil {
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

func validateOutputFormat(verbose bool, jsonFlag bool, xmlFlag bool) {
	var err error
	if jsonFlag == true && xmlFlag == true {
		err = errors.New("Error: Invalid output formats combination provided, enter either xml or json")
	} else if verbose == true && (jsonFlag == true || xmlFlag == true) {
		err = errors.New("Error: Invalid output format combination provided. enter only one of verbose, xml or json formats")
	}
	if err != nil {
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

func validateSeverity(severity string) {
	var err error
	switch severity {
	case tunelinux.All:
	case tunelinux.Severity.String(tunelinux.Info):
	case tunelinux.Severity.String(tunelinux.Warning):
	case tunelinux.Severity.String(tunelinux.Critical):
	case tunelinux.Severity.String(tunelinux.Error):
	default:
		err = errors.New("Error: Invalid recommendation severity provided")
	}
	if err != nil {
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

// handle get recommendations command
func handleGetRecommendations(getCommandOptions *getOptions) (err error) {
	// get recommendations for category
	err = getRecommendationByCategory(getCommandOptions.category)
	if err != nil {
		return err
	}
	// display recommendations
	if getCommandOptions.json == true {
		displayCustomFormat(JSON, getCommandOptions.status, getCommandOptions.severity)
	} else if getCommandOptions.xml == true {
		displayCustomFormat(XML, getCommandOptions.status, getCommandOptions.severity)
	} else if getCommandOptions.verbose == true {
		displayVerboseFormat(getCommandOptions.status, getCommandOptions.severity)
	} else {
		displayTabularFormat(getCommandOptions.status, getCommandOptions.severity)
	}
	return err
}

// handle set recommendations command
func handleSetRecomendations(setCommandOptions *setOptions) (err error) {
	// set recommendations for category
	err = setRecommendationsByCategory(setCommandOptions.category, setCommandOptions.global)
	return err
}

var (
	// Version contains the current version added by the build process
	Version = "dev"
	// Commit containers the hg commit added by the build process
	Commit = "unknown"
)

func getVersion() (version string) {
	return "Version: " + Version + " Commit: " + Commit
}

const (
	getCommandDescription = "Get recommendations"
	setCommandDescription = "Set recommendations"
	categoryDescription   = "Recommendation category {filesystem | multipath | disk | iscsi | fc | All}. (Optional)"
	complianceDescription = "Recommendation status {recommended | not-recommended | All}. (Optional)"
	severityDescription   = "Recommendation severity {critical | warning | info | All}. (Optional)"
	jsonDescription       = "JSON output of recommendations. (Optional)"
	xmlDescription        = "XML output of recommendations. (Optional)"
	verboseDescription    = "Verbose output. (Optional)"
	versionDescription    = "Display version of the tool. (Optional)"
	globalDescription     = "Configure settings at the host level {true | false}. (Optional)"
	NimbleTuneLogFile     = "/var/log/nimbletune.log"
)

var (
	getFlag          = flag.Bool("get", false, getCommandDescription)
	setFlag          = flag.Bool("set", false, setCommandDescription)
	category         = flag.String("category", tunelinux.All, categoryDescription)
	complianceStatus = flag.String("status", tunelinux.All, complianceDescription)
	severity         = flag.String("severity", tunelinux.All, severityDescription)
	verbose          = flag.Bool("verbose", false, verboseDescription)
	jsonFlag         = flag.Bool("json", false, jsonDescription)
	xmlFlag          = flag.Bool("xml", false, xmlDescription)
	versionFlag      = flag.Bool("version", false, xmlDescription)
	global           = flag.Bool("global", false, globalDescription)
)

// initialize command options for short options
func init() {
	flag.StringVar(category, "c", tunelinux.All, categoryDescription)
	flag.StringVar(complianceStatus, "st", tunelinux.All, complianceDescription)
	flag.StringVar(severity, "sev", tunelinux.All, severityDescription)
	flag.BoolVar(versionFlag, "v", false, versionDescription)
}

func main() {
	var err error

	// override Usage
	flag.Usage = func() {
		fmt.Printf("\nNimble Linux Tuning Utility\n")
		fmt.Printf("\nUsage:\n")
		fmt.Println()
		fmt.Printf("nimbletune --get [—category {filesystem | multipath | disk | iscsi | fc | all}] [—status {recommended | not-recommended | all}] [—severity {critical | warning | info | all}] [—verbose] [-json] [-xml]\n")
		fmt.Printf("nimbletune --set [—category {multipath | disk | iscsi | all}]\n")
		fmt.Printf("\nOptions:\n")
		fmt.Printf("\t%-20s\t%-50s\n", "-c, -category", categoryDescription)
		fmt.Printf("\t%-20s\t%-50s\n", "-st, -status", complianceDescription)
		fmt.Printf("\t%-20s\t%-50s\n", "-sev, -severity", severityDescription)
		fmt.Printf("\t%-20s\t%-50s\n", "-json", jsonDescription)
		fmt.Printf("\t%-20s\t%-50s\n", "-xml", xmlDescription)
                fmt.Printf("\t%-20s\t%-50s\n", "-global", globalDescription)
		fmt.Printf("\t%-20s\t%-50s\n", "-verbose", verboseDescription)
		fmt.Printf("\t%-20s\t%-50s\n", "-v, -version", versionDescription)
		fmt.Println()
	}

	log.InitLogging(NimbleTuneLogFile, &log.LogParams{Level: "trace"}, false)

	flag.Parse()

	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if flag.Parsed() != true {
		fmt.Println("Error parsing command options")
		flag.Usage()
		os.Exit(2)
	} else if *versionFlag == true {
		fmt.Println(getVersion())
		return
	} else if *getFlag == false && *setFlag == false {
		fmt.Println("Please pass the get or set subcommand for recommendations")
		flag.Usage()
		os.Exit(1)
	}

	// Check if options get/set suboptions are parsed.
	if *getFlag == true {
		// get the options structure
		getCommandOptions := &getOptions{
			category: *category,
			status:   *complianceStatus,
			severity: *severity,
			verbose:  *verbose,
			json:     *jsonFlag,
			xml:      *xmlFlag}

		// validate input parameters
		validateCategory(*category)
		validateStatus(*complianceStatus)
		validateSeverity(*severity)
		validateOutputFormat(*verbose, *jsonFlag, *xmlFlag)
		// handle get command
		err = handleGetRecommendations(getCommandOptions)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else if *setFlag == true {
		// TODO set recommendations
		// get the options structure
		setCommandOptions := &setOptions{
			category: *category,
			severity: *severity,
			global:   *global}

		validateCategory(*category)
		validateSeverity(*severity)
		// handle get command
		err = handleSetRecomendations(setCommandOptions)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Println("Please pass the get or set subcommand for recommendations")
		os.Exit(1)
	}
}
