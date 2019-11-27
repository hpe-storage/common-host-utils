package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hpe-storage/common-host-libs/chapi"
	"github.com/hpe-storage/common-host-libs/jconfig"
	"github.com/hpe-storage/common-host-libs/model"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var (
	server       *httptest.Server
	hostsURL     string
	hosts        model.Hosts
	devices      []*model.Device
	device       model.Device
	initiators   []*model.Initiator
	networks     []*model.NetworkInterface
	partitions   []*model.DevicePartition
	mounts       []*model.Mount
	uuid         string
	hostResp     []byte
	serialnumber string
	devicesReq   []*model.Device
	initiatorReq []*model.Initiator
	networkReq   []*model.NetworkInterface
	mountStruct  model.Mount
	mountID      string
	config       *jconfig.Config
)

type Config struct {
	MountStr string
}

func init() {
	var err error
	config, err = jconfig.NewConfig("./testconf.json")
	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}
	server = httptest.NewServer(chapi.NewRouter())
	defer server.Close()
	hostsURL = fmt.Sprintf("%s/hosts", server.URL)
	hostuuid := config.GetString("id")
	hostsStruct := model.Hosts{&model.Host{UUID: hostuuid}}
	hostResp, err = json.Marshal(hostsStruct)
	if err != nil {
		fmt.Println(err)
	}
	serialnumber = config.GetString("serialnumber")
	devReq := model.Device{
		Pathname:        config.GetString("pathname"),
		Major:           config.GetString("major"),
		Minor:           config.GetString("minor"),
		AltFullPathName: config.GetString("altfullpathname"),
		SerialNumber:    config.GetString("serialnumber"),
		MpathName:       config.GetString("mpathname"),
	}
	devicesReq = append(devicesReq, &devReq)

	initReq := model.Initiator{
		Type: "iscsi",
		Init: []string{config.GetString("iscsiinitiator")},
	}
	initiatorReq = append(initiatorReq, &initReq)

	netReq := model.NetworkInterface{
		Name:      config.GetString("netinf"),
		AddressV4: config.GetString("addressV4"),
	}
	networkReq = append(networkReq, &netReq)
	mountID = config.GetString("mountid")

	// The datatype is string , conversion is not needed.
	//mountID, err := strconv.ParseUint(mountidstr, 10, 64)

	if err != nil {
		fmt.Print("Unable to retrieve mount id from conf file", mountID)
	}
	mountStruct = model.Mount{
		Mountpoint: config.GetString("mountpoint"),
		ID:         mountID,
		Device: &model.Device{
			Pathname:        config.GetString("pathname"),
			Major:           config.GetString("major"),
			Minor:           config.GetString("minor"),
			AltFullPathName: config.GetString("altfullpathname"),
			SerialNumber:    config.GetString("serialnumber"),
			MpathName:       config.GetString("mpathname"),
		},
	}
}

func TestGetHosts(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(hostResp))
	}
	req := httptest.NewRequest("GET", hostsURL, nil)

	if req.URL.String() != hostsURL {
		t.Error(req.URL.String())
	}

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}
	json.Unmarshal(body, &hosts)
	uuid = hosts[0].UUID
	if !IsValidUUID(hosts[0].UUID) {
		t.Errorf("Valid UUID expected %s", hosts[0].UUID)
	}
}

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func TestGetInitiators(t *testing.T) {
	initiatorResp, err := json.Marshal(initiatorReq)
	if err != nil {
		t.Error(err)
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(initiatorResp))
	}

	req := httptest.NewRequest("GET", hostsURL+uuid+"/initiators", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	buf, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(buf, &initiators)

	if initiators[0].Type != "iscsi" {
		t.Errorf("Expected a iscsi initiator %s", initiators[0].Init)
	}
}

func TestGetNetworks(t *testing.T) {
	networkResp, err := json.Marshal(networkReq)
	if err != nil {
		t.Error(err)
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(networkResp))
	}

	req := httptest.NewRequest("GET", hostsURL+uuid+"/networks", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(buf, &networks)

	if networks[0].Name == "" {
		t.Errorf("Expected a network interface %s", networks[0].Name)
	}
}

func TestGetDevices(t *testing.T) {
	devicesResp, err := json.Marshal(devicesReq)
	if err != nil {
		t.Error(err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(devicesResp))
	}

	req := httptest.NewRequest("GET", hostsURL+uuid+"/devices", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(buf, &devices)

	if devices[0].AltFullPathName == "" || !strings.Contains(devices[0].AltFullPathName, "/dev/mapper") {
		t.Errorf("Expected a device mapper device %s", devices[0].AltFullPathName)
	}
}

func TestCreateDevices(t *testing.T) {
	var vols []*model.Volume
	vol := model.Volume{
		Name:           config.GetString("name"),
		SerialNumber:   config.GetString("serialnumber"),
		AccessProtocol: config.GetString("accessprotocol"),
		Iqn:            config.GetString("iqn"),
		DiscoveryIP:    config.GetString("discoveryip"),
	}
	vols = append(vols, &vol)

	volsReq, err := json.Marshal(vols)
	if err != nil {
		t.Error(err)
	}

	devicesResp, err := json.Marshal(devicesReq)
	if err != nil {
		t.Error(err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(devicesResp))
	}

	req := httptest.NewRequest("POST", hostsURL+uuid+"/devices", bytes.NewBuffer(volsReq))
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	if w.Body == nil {
		t.Errorf("Body = nil, expected some device")
	}
	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}

	buf, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(buf, &devices)

	if devices[0].AltFullPathName == "" || !strings.Contains(devices[0].AltFullPathName, "/dev/mapper") {
		t.Errorf("Expected a device mapper device %s", devices[0].AltFullPathName)
	}
}

func TestDeleteDevice(t *testing.T) {

	iscsiTarget := &model.IscsiTarget{
		Name:    config.GetString("iqn"),
		Address: config.GetString("discoveryip"),
		Port:    "3260",
	}
	deviceToDelete := &model.Device{
		Pathname:        config.GetString("pathname"),
		Major:           config.GetString("major"),
		Minor:           config.GetString("minor"),
		AltFullPathName: config.GetString("altfullpathname"),
		SerialNumber:    config.GetString("serialnumber"),
		MpathName:       config.GetString("mpathname"),
		Slaves:          config.GetStringSlice("path"),
		IscsiTarget:     iscsiTarget,
	}

	deviceReq, err := json.Marshal(deviceToDelete)
	if err != nil {
		t.Error(err)
	}

	deviceResp := deviceReq
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(deviceResp))
	}
	req := httptest.NewRequest("DELETE", hostsURL+uuid+"/devices/"+serialnumber, bytes.NewBuffer(deviceReq))
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(buf, &device)

	if device.AltFullPathName == "" || !strings.Contains(device.AltFullPathName, "/dev/mapper") {
		t.Errorf("Expected a device mapper device %s", device.AltFullPathName)
	}
}

func TestGetPartitions(t *testing.T) {
	var partitionsStruct []*model.DevicePartition

	partStruct := model.DevicePartition{
		Name:          config.GetString("mpathname"),
		Partitiontype: config.GetString("partitiontype"),
		Size:          config.GetInt64("size"),
	}
	partitionsStruct = append(partitionsStruct, &partStruct)

	partitionResp, err := json.Marshal(partitionsStruct)
	if err != nil {
		t.Error(err)
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(partitionResp))
	}
	req := httptest.NewRequest("GET", hostsURL+"/"+uuid+"/devices/"+serialnumber+"/partitions", nil)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(buf, &partitions)

	if partitions[0].Partitiontype != "mpath" || partitions[0].Size <= 0 {
		t.Errorf("Unable to find a valid Partition of a Device %s", partitions[0].Name)
	}
}

func TestGetMounts(t *testing.T) {
	var mountsStruct []*model.Mount
	mountsStruct = append(mountsStruct, &mountStruct)

	mountsResp, err := json.Marshal(mountsStruct)
	if err != nil {
		t.Error(err)
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(mountsResp))
	}
	req := httptest.NewRequest("GET", hostsURL+"/"+uuid+"/mounts", nil)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}
	buf, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(buf, &mounts)

	if mounts[0].Mountpoint == "" {
		t.Errorf("Unable to find a valid mount point of a Device %s", mounts[0].Device.Pathname)
	}

}

func TestCreateFileSystem(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(""))
	}
	req := httptest.NewRequest("PUT", hostsURL+uuid+"/devices/"+serialnumber, nil)
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()
	buf, _ := ioutil.ReadAll(resp.Body)

	if string(buf) != "" {
		t.Errorf("Body =  " + string(buf) + "expected empty body")
	}
	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}
}

func TestCreateMount(t *testing.T) {
	mountReq, err := json.Marshal(mountStruct)
	if err != nil {
		t.Error(err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(mountReq))
	}
	req := httptest.NewRequest("POST", hostsURL+uuid+"/mounts", bytes.NewBuffer(mountReq))
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()
	buf, _ := ioutil.ReadAll(resp.Body)

	var mnt *model.Mount
	json.Unmarshal(buf, &mnt)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}

	if mnt == nil || mnt.Mountpoint == "" {
		t.Errorf("No Mount point created")
	}
}

func TestUnmount(t *testing.T) {
	mountReq, err := json.Marshal(mountStruct)
	if err != nil {
		t.Error(err)
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, string(mountReq))
	}
	req := httptest.NewRequest("DELETE", hostsURL+uuid+"/mounts"+mountID, bytes.NewBuffer(mountReq))
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()
	buf, _ := ioutil.ReadAll(resp.Body)

	var mnt *model.Mount
	json.Unmarshal(buf, &mnt)

	if resp.StatusCode != 200 {
		t.Errorf("Success expected %d", resp.StatusCode)
	}

	if mnt == nil || mnt.Mountpoint == "" {
		t.Errorf("No Mount point created")
	}
}
