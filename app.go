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
	"time"

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



// doAnything is the main app function which is called periodically
func processDevices(configId int64) {
	config, err := conf.GetConfig(context.Background(), configId)
	if err != nil {
		log.Error("devices", "Error reading configuration: %v", err)
		return
	}
	Devices, devicelist, err := fetchDevicesAndSetActiveState(config)
	if err!= nil {
		return
	}
	if config.ProjIds != nil {
		for _, projId := range *config.ProjIds {
			for device := range devicelist.Result {
				_, err := getOrCreateMapping(config, projId, devicelist, device, Devices)
				if err != nil {
					return
				}
			}
		}
	}
	//TODO: Send Data to Eliona
}



func fetchDevicesAndSetActiveState(config *apiserver.Configuration) ([]glutz.DeviceDb,*glutz.DeviceGlutz, error) {
	if config.Enable == nil || !*config.Enable {
		_, err :=conf.SetConfigActiveState(config.ConfigId, false)
		return nil, nil, err
	}
	conf.SetConfigActiveState(config.ConfigId, true)
	var Devices []glutz.DeviceDb
	deviceList, err := GetDevices(config)
	if err != nil {
		return nil, nil, err
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
			OpenableDuration: "",
		}
		Devices = append(Devices, Device)
	}
	return Devices, deviceList, nil
}

func getOrCreateMapping( config *apiserver.Configuration, projId string, devicelist *glutz.DeviceGlutz, device int, Devices []glutz.DeviceDb) (*apiserver.Device, error) {
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

func createAssetandMapping(config *apiserver.Configuration, projId string, deviceid string, assetname string, locationId string)(*apiserver.Device, error){
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
	confDevice, err := conf.GetDevice(context.Background(),config.ConfigId, projId, deviceid)
	if err != nil {
		log.Error("devices", "Error when reading devices from configurations")
		return nil, err
	}
	return confDevice, nil
}


func GetDevices(config *apiserver.Configuration)(*glutz.DeviceGlutz, error){
	deviceRequest := Request{
		Jsonrpc: "2.0",
		ID: "m",
		Method: "eAccess.getModel",
		Params:[]interface{}{
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

func GetDeviceStatus(config *apiserver.Configuration, device_id string)(*glutz.DeviceStatusGlutz, error) {
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

func GetLocation(config *apiserver.Configuration, accessPointId string)(*glutz.DeviceAccessPointGlutz,error){
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


// listenApi starts the API server and listen for requests
func listenApi() {
	http.ListenApiWithOs(&nethttp.Server{Addr: ":" + common.Getenv("API_SERVER_PORT", "3000"), Handler: apiserver.NewRouter(
		apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
		apiserver.NewVersionApiController(apiservices.NewVersionApiService()),
	)})
}
