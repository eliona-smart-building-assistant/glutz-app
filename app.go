//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"context"
	"glutz/apiserver"
	"glutz/apiservices"
	"glutz/conf"
	"glutz/eliona"
	"glutz/glutz"
	nethttp "net/http"
	"strconv"
	"time"
	"fmt"
	

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

type Request struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type DeviceParams struct {
	DeviceID string `json:"deviceid"`
}

type Duration struct {
	Duration string `json:"Duration"`
}

type OutputData struct{
	Openable float64
}

func checkConfigandSetActiveState() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}

	for _, config := range configs {
		// Skip config if disabled and set inactive
		if !conf.IsConfigEnabled(config) {
			if conf.IsConfigActive(config) {
				conf.SetConfigActiveState(config.ConfigId, false)
			}
			continue
		}

		// Signals that this config is active
		if !conf.IsConfigActive(config) {
			conf.SetConfigActiveState(config.ConfigId, true)
			log.Info("conf", "Collecting initialized with Configuration %d:\n"+
				"Username: %s\n"+
				"Password: %s\n"+
				"API Token: %s\n"+
				"Active: %t\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Project IDs: %v\n",
				config.ConfigId,
				config.Username,
				config.Password,
				config.ApiToken,
				*config.Active,
				*config.Enable,
				config.RequestTimeout,
				config.RefreshInterval,
				*config.ProjIds)
		}

		// Runs the ReadNode. If the current node is currently running, skip the execution
		// After the execution sleeps the configured timeout. During this timeout no further
		// process for this config is started to read the data.
		common.RunOnceWithParam(func(config apiserver.Configuration) {
			log.Info("main", "Processing devices for configId %d started", config.ConfigId)

			processDevices(config)

			log.Info("main", "Processing devices for configId %d finished", config.ConfigId)

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, config.ConfigId)
	}
}

func processDevices(config apiserver.Configuration) {
	Devices, devicelist, err := fetchDevicesAndCreateGlutzProperty(config)
	if err != nil {
		return
	}
	if config.ProjIds != nil {
		for _, projId := range *config.ProjIds {
			for device := range devicelist.Result {
				confDevice, err := getOrCreateMapping(config, projId, devicelist, device, Devices)
				if err != nil {
					return
				}
				err = sendData(Devices, device, confDevice)
				if err != nil {
					return
				}
			}
		}
	}
}

func fetchDevicesAndCreateGlutzProperty(config apiserver.Configuration) ([]glutz.DeviceDb, *glutz.DeviceGlutz, error) {
	var Devices []glutz.DeviceDb
	deviceList, err := GetDevices(config)
	if err != nil {
		return nil, nil, err
	}
	openableDurationSet, err := setAccessPointPropertyOpenableDuration(config)
	if err != nil {
		return nil, nil, err
	}
	if openableDurationSet {
		conf.SetConfigInitialisedState(config.ConfigId, true)
	}
	for result := range deviceList.Result {
		deviceid := deviceList.Result[result].Deviceid
		deviceStatus, err := GetDeviceStatus(config, deviceid)
		if err != nil {
			return nil, nil, err
		}
		accesspointid := deviceList.Result[result].AccessPointId
		accessPointId, err := GetLocation(config, accesspointid)
		if err != nil {
			return nil, nil, err
		}
		Device := glutz.DeviceDb{
			BatteryLevel:     deviceStatus.Result[0].BatteryLevel,
			Openings:         deviceStatus.Result[0].Openings,
			Building:         accessPointId.Result[0],
			Room:             accessPointId.Result[1],
			AccessPoint:      accessPointId.Result[2],
			OperatingMode:    deviceStatus.Result[0].OperatingMode,
			Firmware:         deviceStatus.Result[0].Firmware,
		}
		Devices = append(Devices, Device)
	}
	return Devices, deviceList, nil
}

func getOrCreateMapping(config apiserver.Configuration, projId string, devicelist *glutz.DeviceGlutz, device int, Devices []glutz.DeviceDb) (*apiserver.Device, error) {
	confDevice, err := conf.GetDevice(context.Background(), config.ConfigId, projId, devicelist.Result[device].Deviceid)
	if err != nil {
		log.Error("spaces", "Error when reading devices from configurations")
	}
	assetname := Devices[device].AccessPoint + ", " + Devices[device].Room + ", " + Devices[device].Building
	locationid := devicelist.Result[device].AccessPointId
	if confDevice == nil {
		confDevice, err = createAssetandMapping(config, projId, devicelist.Result[device].Deviceid, assetname, locationid)
		if err != nil {
			log.Debug("devices", "Error creating asset and mapping")
			return nil, err
		}
	} else {
		exists, err := asset.ExistAsset(confDevice.AssetId)
		if err != nil {
			log.Error("devices", "Error when checking if asset already exists")
			return nil, err
		}
		if exists {
			log.Debug("devices", "Asset already exists for device %v with AssetId %v", assetname, confDevice.AssetId)
		} else {
			log.Debug("devices", "Asset with AssetId %v does no longer exist in eliona", confDevice.AssetId)
			return nil, nil
		}
	}
	return confDevice, nil
}

func createAssetandMapping(config apiserver.Configuration, projId string, deviceid string, assetname string, locationId string) (*apiserver.Device, error) {
	assetId, err := eliona.CreateNewAsset(projId, deviceid, assetname)
	if err != nil {
		log.Error("devices", "Error when creating new asset")
		return nil, err
	}
	log.Debug("devices", "AssetId %v assigned to device %v", assetId, assetname)
	err = conf.InsertSpace(context.Background(), config.ConfigId, projId, deviceid, assetId, locationId)
	if err != nil {
		log.Error("devices", "Error when inserting device into database:%v", err)
		return nil, err
	}
	log.Debug("devices", "Asset with AssetId %v corresponding to device %v inserted into eliona database", assetId, assetname)
	confDevice, err := conf.GetDevice(context.Background(), config.ConfigId, projId, deviceid)
	if err != nil {
		log.Error("devices", "Error when reading devices from configurations")
		return nil, err
	}
	return confDevice, nil
}

func sendData(Devices []glutz.DeviceDb, device int, confDevice *apiserver.Device) error {
	err := eliona.UpsertInputData(Devices[device], confDevice.AssetId)
	if err != nil {
		return err
	}
	eliona.UpsertInfoData(Devices[device], confDevice.AssetId)
	if err != nil {
		return err
	}
	return nil
}

func GetDevices(config apiserver.Configuration) (*glutz.DeviceGlutz, error) {

	deviceRequest := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.getModel",
		Params: []interface{}{
			"Devices",
		},
	}
	devicerequest, err := http.NewPostRequest(config.Url, deviceRequest)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return nil, err
	}
	deviceList, err := http.Read[glutz.DeviceGlutz](devicerequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading devices: %v", err)
		return nil, err
	}
	return &deviceList, nil
}

func setAccessPointPropertyOpenableDuration(config apiserver.Configuration) (bool, error) {
	req := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.setAccessPointProperty",
		Params: []interface{}{
			"/Properties/Eliona/Openable Duration [s]",
			"",
			"0",
		},
	}
	accesspointrequest, err := http.NewPostRequest(config.Url, req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return false, err
	}
	propertyset, err := http.Read[glutz.Properties](accesspointrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device status: %v", err)
		return false, err
	}
	return propertyset.Result, nil
}


func getAccessPointPropertyOpenableDuration(config apiserver.Configuration, locationid string) (string, error) {
	req := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.getAccessPointProperty",
		Params: []interface{}{
			"/Properties/Eliona/Openable Duration [s]",
			locationid,
		},
	}
	accesspointrequest, err := http.NewPostRequest(config.Url, req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return "", err
	}
	propertyget, err := http.Read[glutz.GlutzOpenableDuration](accesspointrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device status: %v", err)
		return "", err
	}
	return propertyget.Result, nil
}

func sendOpenableDurationToDoor(config apiserver.Configuration, openableDuration int, locationid string)(bool, error){
	//TODO: Improve this and make it more general
	durationstring := "00:00:"
	req := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.openAccessPoint",
		Params: []interface{}{
			locationid,
			Duration{Duration: durationstring + strconv.Itoa(openableDuration)},
		},
	}
	setdurationrequest, err := http.NewPostRequest(config.Url, req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return false, err
	}
	durationset, err := http.Read[glutz.Properties](setdurationrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device status: %v", err)
		return false, err
	}
	return durationset.Result, nil
}


func GetDeviceStatus(config apiserver.Configuration, device_id string) (*glutz.DeviceStatusGlutz, error) {
	req := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.getModel",
		Params: []interface{}{
			"DeviceStatus",
			DeviceParams{DeviceID: device_id},
		},
	}
	devicestatusrequest, err := http.NewPostRequest(config.Url, req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return nil, err
	}
	deviceStatus, err := http.Read[glutz.DeviceStatusGlutz](devicestatusrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device status: %v", err)
		return nil, err
	}
	return &deviceStatus, nil
}

func GetLocation(config apiserver.Configuration, accessPointId string) (*glutz.DeviceAccessPointGlutz, error) {
	req := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.getAccessPointProperty",
		Params: []interface{}{
			"location",
			accessPointId,
		},
	}
	deviceaccesspointrequest, err := http.NewPostRequest(config.Url, req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return nil, err
	}
	deviceAccessPoint, err := http.Read[glutz.DeviceAccessPointGlutz](deviceaccesspointrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device access point: %v", err)
		return nil, err
	}
	return &deviceAccessPoint, nil
}




func checkForOutputChanges() {
	// Generate Connection for listening
	conn, err := http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/data-listener?dataSubtype=output", "X-API-Key", common.Getenv("API_TOKEN", ""))
	if err != nil {
		log.Error("Output", "Error creating web socket connection")
		return
	}
	// Any changes in database written to outputs channel
	outputs := make(chan api.Data)
	go http.ListenWebSocket(conn, outputs)
	for output := range outputs {
		// Check the assetid of the updated entry is a Glutz device
		DeviceExists, err := conf.ExistGlutzDeviceWithAssetId(context.Background(), output.AssetId)
		if err != nil {
			log.Error("Output", "Error checking if asset id corresponds to a glutz device")
			return
		}
		data, _ := mapToStruct(output.Data)
		openable:= data.Openable
		if DeviceExists && openable == 1 {
			// Fetch the Glutz device where a value was changed in the database
			device, err := conf.GetDevicewithAssetId(context.Background(), output.AssetId)
			if err != nil {
				log.Error("Output", "Error getting device from assetid %v", err)
				return
			}
			config, err:= conf.GetConfig(context.Background(), int64(device.ConfigId))
			if err!= nil {
				log.Error("Output", "Error getting configuration %v", err)
				return
			}
			// Check if a value exists in glutz environment for openable duration for this door. If so, use this value.
			glutzOpenableDuration, err := getAccessPointPropertyOpenableDuration(*config, device.LocationId)
			if err!= nil {
				log.Error("Output", "Error sending openable duration to door")
				return
			}
			var openableDuration int
			if glutzOpenableDuration != ""  {
				openableDuration, err = strconv.Atoi(glutzOpenableDuration)
				if err != nil {
					log.Error("Output", "Couldn't convert to integer %v", err)
					return
				}
			// If not, use the default value from the config table
			}else{
				openableDuration = int(config.DefaultOpenableDuration)
			}
			//Send Openable Duration to Door
			response, err := sendOpenableDurationToDoor(*config, int(openableDuration), device.LocationId)
			if err!= nil {
				log.Error("Output", "Error sending openable duration to door")
				return
			}
			if response{
				log.Debug("Output", "Opened door at Location %v for %v seconds",device.LocationId, openableDuration)
			}
			
		}
	}
}

func mapToStruct(m map[string]interface{}) (*OutputData, error) {
    s := &OutputData{}

    if v, ok := m["openable"].(float64); ok {
        s.Openable = v
    } else {
        return nil, fmt.Errorf("invalid type for field 'openable'")
    }

    return s, nil
}

// listenApi starts the API server and listen for requests
func listenApi() {
	http.ListenApiWithOs(&nethttp.Server{Addr: ":" + common.Getenv("API_SERVER_PORT", "3000"), Handler: apiserver.NewRouter(
		apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
		apiserver.NewVersionApiController(apiservices.NewVersionApiService()),
	)})
}
