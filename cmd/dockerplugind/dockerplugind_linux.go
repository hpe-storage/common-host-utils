package main

import (
	"github.com/hpe-storage/common-host-libs/chapi"
	"github.com/hpe-storage/common-host-libs/dockerplugin"
	log "github.com/hpe-storage/common-host-libs/logger"
	"os"
	"os/signal"
	"syscall"
)

// InvokePluginDaemon create local and global channel and invoke RunNimbledockerd
func InvokePluginDaemon() {
	pluginChan := make(chan error)
	nimbledChan := make(chan error)

	// Run chapid
	chapi.RunNimbled(nimbledChan)
	// Run docker plugin
	err := dockerplugin.RunNimbledockerd(pluginChan, Version)
	if err != nil {
		log.Fatalf("unable to run docker plugin daemon, err %v", err.Error())
		os.Exit(1)
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		log.Fatalf("Exiting due to signal notification. Signal was %v.", s.String())
		os.Exit(1)
	}()
	select {
	case msg := <-nimbledChan:
		log.Trace("error on chapid socket", msg)
	case msg := <-pluginChan:
		log.Trace("error on hpe docker plugin socket", msg)
	}
	return
}
