package main

import (
	"errors"
	"fmt"
	"github.com/hpe-storage/common-host-libs/chapi"
	"github.com/hpe-storage/common-host-libs/dockerplugin"
	"github.com/hpe-storage/common-host-libs/dockerplugin/plugin"
	"github.com/hpe-storage/common-host-libs/dockerplugin/provider"
	"github.com/hpe-storage/common-host-libs/linux"
	log "github.com/hpe-storage/common-host-libs/logger"
	"github.com/hpe-storage/common-host-libs/tunelinux"
	"github.com/hpe-storage/common-host-libs/util"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

const (
	envIP                    = "PROVIDER_IP"
	envUsername              = "PROVIDER_USERNAME"
	envPassword              = "PROVIDER_PASSWORD"
	envRemove                = "PROVIDER_REMOVE"
	multipathConfPath        = "/etc/multipath.conf"
	stagingConfigPath        = "/opt/hpe-storage/etc/"
	stagingFlexVolumeBinPath = "/opt/hpe-storage/flexvolume"
	doryConfigFile           = "dory.json"
	nimbleFlexVolPath        = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/hpe.com~nimble/"
	hpecvFlexVolPath         = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/hpe.com~cv/"
	simplivityFlexVolPath    = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/hpe.com~simplivity/"
	cleanupHookPath          = "/etc/hpe-storage/remove"
)

var (
	// Version :
	Version = "dev"
	// Commit :
	Commit = "unknown"
)

func main() {
	// initialize logging with defaults
	log.InitLogging(plugin.PluginLogFile, nil, true)

	// check if plugin is called to just remove certs and exit here
	if os.Getenv(envRemove) == "true" {
		err := cleanup()
		if err != nil {
			log.Errorf("unable to cleanup plugin config files and array certificates, err %s", err.Error())
		}
		return
	}

	// configure pre-requisites for plugin
	err := configure()
	if err != nil {
		log.Fatalf("unable to configure docker volume plugin, err %s", err.Error())
	}

	nimbledChan := make(chan error)
	dockerpluginChan := make(chan error)
	// Run chapid
	chapi.RunNimbled(nimbledChan)
	// Run plugin
	err = dockerplugin.RunNimbledockerd(dockerpluginChan, Version)
	if err != nil {
		log.Fatalf("unable to run docker plugin daemon, err %v", err.Error())
	}
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		log.Errorf("Exiting due to signal notification.  Signal was %v.", s.String())
		// check if we are indicated to remove array certificates as part of plugin exit
		// this will be set from preStop hook for flexvolume plugin(daemonset)
		exists, _, _ := util.FileExists(cleanupHookPath)
		if exists {
			// set remove env, so that certificates can be removed
			os.Setenv(envRemove, "true")
			err := cleanup()
			if err != nil {
				log.Errorf("unable to cleanup plugin config files and array certificates during termination of plugin, err %s", err.Error())
			}
		}
		return
	}()
	select {
	case msg := <-nimbledChan:
		log.Error("error on chapid socket:", msg)
	case msg := <-dockerpluginChan:
		log.Error("error on docker plugin socket:", msg)
	}
	// cleanup dory binary copied as well
	err = os.RemoveAll("/usr/libexec/kubernetes/kubelet-plugins/volume/exec/hpe.com~" + plugin.GetPluginType().String())
	if err != nil {
		log.Errorf("unable to cleanup dory on the host, err %s", err.Error())
	}
	log.Info("Successfully cleaned up plugin config files, dory and array certificates")
}

func cleanup() error {
	// cleanup array certificates
	err := handleCertificates()
	if err != nil {
		return err
	}
	// cleanup certificate and config directory on host
	err = os.RemoveAll(plugin.ConfigBaseDir)
	if err != nil {
		log.Errorf("unable to cleanup plugin config directory, err %s", err.Error())
	}
	// cleanup dory binary copied as well
	err = os.RemoveAll("/usr/libexec/kubernetes/kubelet-plugins/volume/exec/hpe.com~" + plugin.GetPluginType().String())
	if err != nil {
		log.Errorf("unable to cleanup dory on the host, err %s", err.Error())
	}
	log.Info("Successfully cleaned up plugin config files, dory and array certificates")
	return nil
}

func configure() (err error) {
	log.Trace(">>> configure")
	defer log.Trace("<<< configure")

	switch plugin.GetPluginType() {
	case plugin.Nimble:
		log.Infof("Starting Nimble Docker Volume plugin  version %s(%s)...", Version, Commit)
		// handle on premise nimble plugin initialization
		err = initializeNimblePlugin()
	case plugin.Simplivity:
		log.Infof("Starting HPE SimpliVity docker volume plugin  version %s(%s)...", Version, Commit)
		// handle simplivity plugin initilization
		err = initializeSimplivityPlugin()
	case plugin.Cv:
		log.Infof("Starting HPE Cloud Volumes docker volume plugin  version %s(%s)...", Version, Commit)
		// handle cloud plugin initialization
		err = initilizeHpeCvPlugin()
	default:
		log.Errorln("unable to determine hpe plugin type to initialize, plugin_type env is missing")
		err = errors.New("unable to determine hpe plugin type to initialize, plugin_type env is missing")
	}
	return err
}

func initializeNimblePlugin() (err error) {
	log.Trace(">>> initializeNimblePlugin")
	defer log.Trace("<<< initializeNimblePlugin")

	// perform service/multipath checks
	err = handleConformance()
	if err != nil {
		return err
	}
	// copy volume-driver.json file
	err = copyDriverConfigFile(plugin.Nimble)
	if err != nil {
		return err
	}
	// handle certificates
	err = handleCertificates()
	if err != nil {
		return err
	}

	// copy flexvolume binary
	err = copyFlexVolumeDriver(plugin.Nimble)
	if err != nil {
		return err
	}
	return nil
}

func initilizeHpeCvPlugin() (err error) {
	log.Trace(">>> initilizeHpeCvPlugin")
	defer log.Trace("<<< initilizeHpeCvPlugin")

	// perform service/multipath checks
	err = handleConformance()
	if err != nil {
		return err
	}
	// copy volume-driver.json file
	err = copyDriverConfigFile(plugin.Cv)
	if err != nil {
		return err
	}
	// copy flexvolume binary
	err = copyFlexVolumeDriver(plugin.Cv)
	if err != nil {
		return err
	}
	return nil
}

func initializeSimplivityPlugin() (err error) {
	log.Trace(">>> initializeSimplivityPlugin")
	defer log.Trace("<<< initializeSimplivityPlugin")

	// copy volume-driver.json file
	err = copyDriverConfigFile(plugin.Simplivity)
	if err != nil {
		return err
	}
	// copy flexvolume binary
	err = copyFlexVolumeDriver(plugin.Simplivity)
	if err != nil {
		return err
	}
	return nil
}

func handleConformance() (err error) {
	// package conformance checks and automatic package installation is not supported for managed plugins
	// so, only copy multipath.conf if missing
	if plugin.IsManagedPlugin() {
		// multipath checks
		if exists, _, _ := util.FileExists(linux.MultipathConf); !exists {
			// Copy the multipath.conf supplied with utility
			multipathTemplate, err := tunelinux.GetMultipathTemplateFile()
			if err != nil {
				return err
			}
			err = util.CopyFile(multipathTemplate, linux.MultipathConf)
			if err != nil {
				return err
			}
		}
		// iscsi checks
		err = tunelinux.SetIscsiRecommendations()
		if err != nil {
			return err
		}
		return nil
	}
	// conformance checks and service management for daemonset
	// configure iscsi
	err = tunelinux.ConfigureIscsi()
	if err != nil {
		return err
	}

	// configure multipath
	err = tunelinux.ConfigureMultipath()
	if err != nil {
		return err
	}
	return nil
}

func copyDriverConfigFile(pluginType plugin.PluginType) (err error) {
	log.Trace(">>> copyDriverConfigFile")
	defer log.Trace("<<< copyDriverConfigFile")

	// for flexvolume plugin config file is provided as config-map in k8s
	if !plugin.IsManagedPlugin() {
		log.Traceln("skip copying flexvolume driver config file as it will be using from configmap")
		return nil
	}

	// check and create the plugin config directory
	pluginConfigDir, err := plugin.GetOrCreatePluginConfigDirectory()
	if err != nil {
		return err
	}

	// check and copy the plugin config file from install location
	exists, _, err := util.FileExists(pluginConfigDir + plugin.DriverConfigFile)
	if !exists {
		pluginConfFile := fmt.Sprintf("%s%s-%s", stagingConfigPath, pluginType.String(), plugin.DriverConfigFile)
		err := util.CopyFile(pluginConfFile, pluginConfigDir+plugin.DriverConfigFile)
		if err != nil {
			return fmt.Errorf("unable to copy driver config %s: error %s", pluginConfigDir+plugin.DriverConfigFile, err.Error())
		}
		log.Infof("successfully copied driver config file as %s", pluginConfigDir+plugin.DriverConfigFile)
	}
	return nil
}

func handleCertificates() (err error) {
	log.Trace(">>> handleCertificates")
	defer log.Trace("<<< handleCertificates")

	// check if certificates are already present
	if certsPresent() {
		// check if we are asked to remove them
		if os.Getenv(envRemove) != "true" {
			return nil
		}
		// verify if required env params are provided to remove certs
		if os.Getenv(envIP) == "" || os.Getenv(envUsername) == "" || os.Getenv(envPassword) == "" {
			return errors.New("missing required environment params username, password and ip to remove Nimble array certificates")
		}
		// remove certificates
		err = removeCertificates(os.Getenv(envIP), os.Getenv(envUsername), os.Getenv(envPassword))
		if err != nil {
			return err
		}
	} else {
		// docker will attempt to restart the plugin multiple times when remove is set, so don't create certs with remove set
		if os.Getenv(envRemove) == "true" {
			log.Infof("plugin is enabled with PROVIDER_REMOVE set, skipping certificate creation")
			return nil
		}
		// create certificates
		err = createCertificates(os.Getenv(envIP), os.Getenv(envUsername), os.Getenv(envPassword))
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFlexVolumeDriver(pluginType plugin.PluginType) error {
	log.Trace(">>> copyFlexVolumeDriver")
	defer log.Trace("<<< copyFlexVolumeDriver")
	var driverPath string

	if plugin.IsManagedPlugin() {
		log.Traceln("managed plugin doesn't support dory/doryd, skip copying flexvolume driver...")
		return nil
	}

	switch pluginType {
	case plugin.Nimble:
		driverPath = nimbleFlexVolPath
	case plugin.Cv:
		driverPath = hpecvFlexVolPath
	case plugin.Simplivity:
		driverPath = simplivityFlexVolPath
	}
	_, isDir, _ := util.FileExists(driverPath)
	if !isDir {
		log.Tracef("%s doesn't exist on the host creating it", driverPath)
		// create the directory on the host
		err := os.MkdirAll(driverPath, 0700)
		if err != nil {
			return err
		}
	}
	// remove existing flexvolume exec file
	os.Remove(driverPath + pluginType.String())

	// copy the binary to the flexvolume driver path
	log.Tracef("copying the %s flexvolume binary to %s", pluginType.String(), driverPath)
	if err := util.CopyFile(stagingFlexVolumeBinPath, driverPath+pluginType.String()); err != nil {
		return fmt.Errorf("unable to copy the flexvolume binary on the host %s", err.Error())
	}
	if err := os.Chmod(driverPath+pluginType.String(), 0744); err != nil {
		return fmt.Errorf("unable to set exec permissions for flexvolume binary on the host %s", err.Error())
	}

	// copy config file dory.json, if not already present
	targetConfigFile := driverPath + pluginType.String() + ".json"
	exists, _, _ := util.FileExists(targetConfigFile)
	if !exists {
		log.Tracef("copying the %s flexvolume config file to %s", pluginType.String(), driverPath)
		configFile := fmt.Sprintf("%s%s-%s", stagingConfigPath, pluginType.String(), doryConfigFile)
		if err := util.CopyFile(configFile, targetConfigFile); err != nil {
			return fmt.Errorf("unable to copy the flexvolume config file on the host %s", err.Error())
		}
	}
	return nil
}

func certsPresent() (isPresent bool) {
	isHostCertFilePresent, _, _ := util.FileExists(provider.HostCertFile)
	isHostKeyFilePresent, _, _ := util.FileExists(provider.HostKeyFile)
	isServerCertFilePresent, _, _ := util.FileExists(provider.ServerCertFile)
	if isHostCertFilePresent && isHostKeyFilePresent && isServerCertFilePresent {
		return true
	}
	return false
}

func removeCertificates(ip string, username string, password string) (err error) {
	log.Tracef(">>> removeCertificates called with ipaddress: %s username %s", ip, username)
	defer log.Trace("<<< removeCertificates")

	present, _, _ := util.FileExists(provider.HostCertFile)
	if !present {
		// certificates not present
		return nil
	}
	hostCertPem, err := ioutil.ReadFile(provider.HostCertFile)
	if err != nil {
		log.Error("unable to load host cert", err.Error())
		return err
	}
	// invoke dockerplugin.NimbleRemoveURI end point of container provider to remove certificate
	err = provider.AddRemoveCertContainerProvider(provider.NimbleRemoveURI, ip, string(hostCertPem), username, password)
	if err != nil {
		return err
	}

	//remove the cert files
	os.RemoveAll(provider.ServerCertFile)
	os.RemoveAll(provider.HostCertFile)
	os.RemoveAll(provider.HostKeyFile)
	return nil
}

func createCertificates(ip string, username string, password string) (err error) {
	log.Trace(">>> createCertificates")
	defer log.Trace("<<< createCertificates")

	log.Tracef("createCertificates called with ipaddress: %s, username: %s", ip, username)
	// create and add certificate to the group
	err = provider.LoginAndCreateCerts(ip, username, password, true)
	if err != nil {
		return fmt.Errorf("unable to login to array and generate certificates, err %s", err.Error())
	}
	return nil
}
