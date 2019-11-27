package main

// Copyright 2019 Hewlett Packard Enterprise Development LP.
import (
	"github.com/hpe-storage/common-host-libs/linux"
	log "github.com/hpe-storage/common-host-libs/logger"
	"github.com/hpe-storage/common-host-libs/model"
	"github.com/hpe-storage/common-host-libs/util"
)

var (
	connectVolLog = util.GetNltHome() + "log/connectvoliscsi.log"
)

// tested this against yhwang-3

//TODO modify this to use testing framework for Golang
func main() {
	log.InitLogging(connectVolLog, &log.LogParams{Level: "trace"}, false)

	discoveryip := "172.16.234.48"
	iqn := "iqn.2007-11.com.nimblestorage:createtest0005-v0bb3b16fff385f8a.0000024a.f48beee5"
	log.Trace("Perform Iscsi Discovery")
	iscsiTargets, err := linux.PerformDiscovery([]string{discoveryip})
	if err == nil {
		for _, s := range iscsiTargets {
			log.Trace(s.Name)
		}
	}

	log.Trace("Get Targets")
	iscsiTargets2, err := linux.GetIscsiTargets()
	if err == nil {
		for _, s := range iscsiTargets2 {
			log.Trace(s.Name)
		}
	}

	log.Trace("Rescan and Login")
	volume := &model.Volume{DiscoveryIP: discoveryip, Iqn: iqn, ConnectionMode: "automatic", LunID: "0"}
	err = linux.RescanAndLoginToTarget(volume)
	if err != nil {
		log.Trace(err.Error())
	}

	log.Trace("Get Nimble Dm Device")
	listDevices, err := linux.GetNimbleDmDevices(false, "", "")
	if err == nil {
		for _, s := range listDevices {
			log.Trace("Device", "path:", s.Pathname)
			log.Trace("Device", "altPath", s.AltFullPathName)
			log.Trace("GetPartitionInfo")
			deviceParitionInfos, err := linux.GetPartitionInfo(s)
			if err != nil {
				log.Trace(err.Error())
				return
			}
			for _, d := range deviceParitionInfos {
				log.Trace("PartInfo: ", "Name:", d.Name, "Size:", d.Size)
			}
		}
	}
	showPathsMaps()
	log.Trace("ReadFirstLineFromFile")
	util.FileReadFirstLine("/sys/block/dm-23/dm/uuid")
	linux.GetMountPointsForDevices(listDevices)
}

func showPathsMaps() {
	mpathshowpaths, err := linux.MultipathdShowPaths("")
	if err == nil {
		for _, s := range mpathshowpaths {
			log.Trace(s)
		}
	}

	log.Trace("Multipathd Show Maps")
	mpathshowmaps, err := linux.MultipathdShowMaps("")
	if err == nil {
		for _, s := range mpathshowmaps {
			log.Trace(s)
		}
	}
}
