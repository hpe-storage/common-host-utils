// Copyright 2017 Nimble Storage, Inc.

// This is a windows service utility which would registered the
// dockerplugind as service on Windows.
// "usage: %s <command> "
//      where <command> is one of "
//      install, remove, debug, start, stop, pause or continue."
//
package main

import (
	"flag"

	"github.com/kardianos/service"
	"github.com/hpe-storage/common-host-libs/dockerplugin"
	log "github.com/hpe-storage/common-host-libs/logger"
	"os"
	"os/signal"
	"syscall"
)

type program struct{}

// Initialize the service attributes.
var svcConfig = &service.Config{
	Name:        "dockerplugind",
	DisplayName: "HPE Nimble Docker Volume Plugin",
	Description: "HPE Docker Volume Plugin Service",
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	// Do work here
	log.Infof("Run dockerplugind.")
	pluginChan := make(chan error)
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
	case msg := <-pluginChan:
		log.Trace("error on plugin listening port:", msg)
	}

}
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	log.Infof("Stopping HPE Docker Volume Plugin Service.")

	return nil
}

//InvokePluginDaemon converts the dockerplugind exe to windows service.
func InvokePluginDaemon() {
	// Command line flag parsing for receiving arguments value.
	flag.Parse()

	prg := &program{}
	nsvc, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalf("unable to setup new service, err %v", err.Error())
	}

	if len(flag.Args()) > 0 {
		err = service.Control(nsvc, flag.Args()[0])
		if err != nil {
			log.Fatalf("unable to setup new service, err %v", err.Error())
		}
		return
	}
	nsvc.Run()
}
