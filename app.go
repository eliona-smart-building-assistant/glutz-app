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
	"glutz/glutz"
	nethttp "net/http"
	"time"
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
	//// Fetch Devices ////
	config, err := conf.GetConfig(context.Background(), configId)
	if err != nil {
		log.Error("devices", "Error reading configuration: %v", err)
	}
	// accesspointId, err := GetLocation(config, "996")
	// if err != nil {
	// 	return
	// }
	// log.Debug("devices","Accesspoint: %v", accesspointId)
	//Get Array of type DeviceDb with all devices found 
	var Devices []glutz.DeviceDb

	deviceList, err := GetDevices(config)
	if err != nil {
		return
	}

	for result := range deviceList.Result {
		deviceid:= deviceList.Result[result].Deviceid
		deviceStatus, err := GetDeviceStatus(config, deviceid)
		if err != nil {
			return
		}
		// accesspointid:=deviceList.Result[result].AccessPointId
		// accessPointId, err := GetLocation(config, accesspointid)
		// if err != nil {
		// 	return
		// }
		Device := glutz.DeviceDb{
			BatteryLevel: deviceStatus.Result[0].BatteryLevel,
			Openings: deviceStatus.Result[0].Openings,
			Building: "examplebuilding",//accessPointId.Result[0].Building,
			Room: "exampleroom",//accessPointId.Result[0].Room,
			AccessPoint: "exampleaccesspoint", //accessPointId.Result[0].AccessPoint,
			OperatingMode: deviceStatus.Result[0].OperatingMode,
			Firmware: deviceStatus.Result[0].Firmware,
			OpenableDuration: "", //Change later
		}
		Devices = append(Devices, Device)
	}
	log.Debug("Devices", "Devices: %v", Devices)

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
