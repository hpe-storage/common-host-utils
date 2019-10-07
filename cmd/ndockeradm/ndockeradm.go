// Copyright 2019 Hewlett Packard Enterprise Development LP

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hpe-storage/common-host-libs/dockerplugin/plugin"
	"github.com/hpe-storage/common-host-libs/dockerplugin/provider"
	log "github.com/hpe-storage/common-host-libs/logger"
)

var (
	// Version contains the current version added by the build process
	Version = "dev"
	// Commit containers the hg commit added by the build process
	Commit = "unknown"
)

func main() {
	// Subcommands
	addGroupCmd := flag.NewFlagSet("add", flag.ExitOnError)
	removeGroupCmd := flag.NewFlagSet("remove", flag.ExitOnError)

	// Add Group subcommand flag pointers
	ipAddressAdd := addGroupCmd.String("ipaddress", "", "GROUP MANAGEMEMT IP")
	username := addGroupCmd.String("username", "", "GROUP USERNAME.")
	password := addGroupCmd.String("password", "", "GROUP PASSWORD")

	// Remove Group subcommand flag pointers
	ipAddressRemove := removeGroupCmd.String("ipaddress", "", "GROUP MANAGEMEMT IP.")
	usernameRemove := removeGroupCmd.String("username", "", "GROUP USERNAME.")
	passwordRemove := removeGroupCmd.String("password", "", "GROUP PASSWORD")

	// override Usage
	flag.Usage = func() {
		fmt.Printf("\nHPE Nimble Storage Docker Admin Utility\n")
		fmt.Printf("\nUsage:\n")
		fmt.Println()
		fmt.Printf("ndockeradm add [-ipaddress {GROUP MANAGEMEMT IP}] [-username {GROUP USERNAME}] [-password {GROUP PASSWORD}] \n")
		fmt.Printf("ndockeradm remove [-ipaddress {GROUP MANAGEMEMT IP}] [-username {GROUP USERNAME}] [-password {GROUP PASSWORD}] \n")
		fmt.Println()
	}
	// check if the necessary logs file directories are created. create if not present
	err := plugin.CreateConfDirectory(DockerLogHome)
	if err != nil {
		log.Error("error to create " + DockerLogHome + " " + err.Error())
		fmt.Printf("\nerror to create " + DockerLogHome + " " + err.Error())
		os.Exit(1)
	}
	// Create certs folder
	err = plugin.CreateConfDirectory(DockerCertHome)
	if err != nil {
		log.Error("error to create " + DockerCertHome + " " + err.Error())
		fmt.Printf("\nerror to create " + DockerCertHome + " " + err.Error())
		os.Exit(1)
	}
	log.InitLogging(DockerLogFile, nil, false)
	log.Infof("ndockeradm version %s(%s) ...", Version, Commit)
	flag.Parse()

	// Verify that a subcommand has been provided
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	// Switch on the subcommand
	switch os.Args[1] {
	case "add":
		addGroupCmd.Parse(os.Args[2:])
	case "remove":
		removeGroupCmd.Parse(os.Args[2:])
	default:
		flag.Usage()
		os.Exit(1)
	}

	if addGroupCmd.Parsed() {
		parseAddGroup(*ipAddressAdd, *username, *password, addGroupCmd)
	} else if removeGroupCmd.Parsed() {
		parseRemoveGroup(*ipAddressRemove, *usernameRemove, *passwordRemove, removeGroupCmd)
	}
}

func parseAddGroup(ipAddressAdd string, username string, password string, addGroupCmd *flag.FlagSet) {
	log.Trace("parseAddGroup called with ", ipAddressAdd, username)
	// Required Flags
	if ipAddressAdd == "" || username == "" || password == "" {
		addGroupCmd.PrintDefaults()
		os.Exit(1)
	}
	log.Tracef("ipaddress: %s, username: %s\n", ipAddressAdd, username)
	// set environment variable for the container-storage-provider ip
	os.Setenv("PROVIDER_IP", ipAddressAdd)
	os.Setenv("PROVIDER_USERNAME", username)
	// create and add certificate to the group
	err := provider.LoginAndCreateCerts(ipAddressAdd, username, password, false)
	if err != nil {
		log.Error(err.Error())
		fmt.Printf(err.Error())
		// delete the invalid certs
		os.RemoveAll(DockerCertHome)
		return
	}
}

func parseRemoveGroup(ipAddressRemove string, username string, password string, removeGroupCmd *flag.FlagSet) {
	log.Trace("parseRemoveGroup called with ", ipAddressRemove)
	if ipAddressRemove == "" || username == "" || password == "" {
		removeGroupCmd.PrintDefaults()
		os.Exit(1)
	}
	log.Tracef("ipaddress: %s\n", ipAddressRemove)
	hostCertPem, err := ioutil.ReadFile(provider.HostCertFile)
	if err != nil {
		log.Trace("unable to load host cert", err.Error())
		fmt.Printf(err.Error())
		return
	}
	log.Tracef("hostCertPem %s", string(hostCertPem))
	// invoke dockerplugin.NimbleRemoveURI end point of container provider to remove certificate
	provider.AddRemoveCertContainerProvider(provider.NimbleRemoveURI, ipAddressRemove, string(hostCertPem), username, password)

	//unset env for container-storage-provider ip
	os.Unsetenv("PROVIDER_IP")
	os.Unsetenv("PROVIDER_USERNAME")

	//remove the cert files
	os.RemoveAll(provider.ServerCertFile)
	os.RemoveAll(provider.HostCertFile)
	os.RemoveAll(provider.HostKeyFile)
}
