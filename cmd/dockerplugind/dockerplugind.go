// Copyright 2019 Hewlett Packard Enterprise Development LP.

package main

import (
	"github.com/hpe-storage/common-host-libs/dockerplugin/plugin"
	log "github.com/hpe-storage/common-host-libs/logger"
)

var (
	// Version contains the current version added by the build process
	Version = "dev"
	// Commit containers the hg commit added by the build process
	Commit = "unknown"
)

// The Main method creates socket for both local and global
//TODO: add dockerplugind_test once the structure is finalized
func main() {
	log.InitLogging(plugin.PluginLogFile, nil, false)

	log.Infof("Starting nimble docker service version %s(%s)...", Version, Commit)
	// This function would actually invoke the daemon
	InvokePluginDaemon()
	return
}
