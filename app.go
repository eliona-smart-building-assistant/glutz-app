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
	"fmt"
	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-eliona/dashboard"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"glutz/apiserver"
	"glutz/apiservices"
	"glutz/conf"
	"glutz/eliona"
	"glutz/glutz"
	nethttp "net/http"
	"strconv"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/gorilla/websocket"
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

type OutputData struct {
	Open float64
}

func initialization() {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, app.AppName())
	defer conn.Close(ctx)

	// Init the app before the first run.
	app.Init(conn, app.AppName(),
		asset.InitAssetTypeFile("eliona/asset-type-glutz_device.json"),
		dashboard.InitWidgetTypeFile("eliona/widget-type-glutz.json"),
		app.ExecSqlFile("conf/init.sql"),
	)
}

func checkConfigAndSetActiveState() {
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
			log.Info("conf", "Collecting initialized with Configuration %d", config.ConfigId)
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
			BatteryLevel:  deviceStatus.Result[0].BatteryLevel,
			Openings:      deviceStatus.Result[0].Openings,
			Building:      accessPointId.Result[0],
			Room:          accessPointId.Result[1],
			AccessPoint:   accessPointId.Result[2],
			OperatingMode: deviceStatus.Result[0].OperatingMode,
			Firmware:      deviceStatus.Result[0].Firmware,
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

// Upserts Input and Info Data to Eliona
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

// Glutz API request to get all devices in configuration
func GetDevices(config apiserver.Configuration) (*glutz.DeviceGlutz, error) {

	deviceRequest := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.getModel",
		Params: []interface{}{
			"Devices",
		},
	}
	devicerequest, err := http.NewPostRequest(config.Url+"/rpc", deviceRequest)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return nil, err
	}
	devicerequest.Header.Add("Referer", config.Url)
	devicerequest.SetBasicAuth(config.Username, config.Password)
	deviceList, err := http.Read[glutz.DeviceGlutz](devicerequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading devices: %v", err)
		return nil, err
	}
	return &deviceList, nil
}

// Glutz API request to initialize the access point property "openable duration" on the Glutz server
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
	accesspointrequest, err := http.NewPostRequest(config.Url+"/rpc", req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return false, err
	}
	accesspointrequest.Header.Add("Referer", config.Url)
	accesspointrequest.SetBasicAuth(config.Username, config.Password)
	propertyset, err := http.Read[glutz.Properties](accesspointrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error setting access point property: %v", err)
		return false, err
	}
	return propertyset.Result, nil
}

// Glutz API request to get the value of the accesspoint property "openable duration" from the Glutz server
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
	accesspointrequest, err := http.NewPostRequest(config.Url+"/rpc", req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return "", err
	}
	accesspointrequest.Header.Add("Referer", config.Url)
	accesspointrequest.SetBasicAuth(config.Username, config.Password)
	propertyget, err := http.Read[glutz.GlutzOpenableDuration](accesspointrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device status: %v", err)
		return "", err
	}
	return propertyget.Result, nil
}

// Glutz API request to get device status of a specific Glutz device
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
	devicestatusrequest, err := http.NewPostRequest(config.Url+"/rpc", req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return nil, err
	}
	devicestatusrequest.Header.Add("Referer", config.Url)
	devicestatusrequest.SetBasicAuth(config.Username, config.Password)
	deviceStatus, err := http.Read[glutz.DeviceStatusGlutz](devicestatusrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device status: %v", err)
		return nil, err
	}
	return &deviceStatus, nil
}

// Glutz API request to get device status of a specific Glutz device
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
	deviceaccesspointrequest, err := http.NewPostRequest(config.Url+"/rpc", req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return nil, err
	}
	deviceaccesspointrequest.Header.Add("Referer", config.Url)
	deviceaccesspointrequest.SetBasicAuth(config.Username, config.Password)
	deviceAccessPoint, err := http.Read[glutz.DeviceAccessPointGlutz](deviceaccesspointrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device access point: %v", err)
		return nil, err
	}
	return &deviceAccessPoint, nil
}

// Generates a websocket connection to the database and listens for any updates on assets (only output attributes). For any update written to the channel
// the function checks whether the assetid of the update is associated with a glutz device and opens it. After the "openable duration" time is up, the door
// is closed again. If the door is currently open, a request to open it again will be ignored.
func listenForOutputChanges() {
	outputs := make(chan api.Data)
	go http.ListenWebSocketWithReconnect(func() (*websocket.Conn, error) {
		return http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/data-listener?dataSubtype=output", "X-API-Key", common.Getenv("API_TOKEN", ""))
	}, 50*time.Millisecond, outputs)
	for output := range outputs {
		openableDoor, _ := checkThereIsADoorToBeOpened(output)
		if openableDoor {
			device, config, _ := getDeviceAndGetConfig(output)
			if device != nil && config != nil {
				openableDuration, _ := getOpenableDuration(config, device)
				if openableDuration > 0 {
					response, _ := sendOpenableDurationToDoor(*config, int(openableDuration), device.LocationId)
					if response {
						err := eliona.UpsertOpenData(1, device.AssetId)
						if err != nil {
							return
						}
						log.Debug("Output", "Opened door at Location %v for %v seconds", device.LocationId, openableDuration)
						go waitAndResetOpen(*config, int(openableDuration), device.AssetId, device.LocationId)
					}
					if !response {
						log.Debug("Output", "Could not open door at Location %v for %v seconds", device.LocationId, openableDuration)
						err := eliona.UpsertOpenData(2, device.AssetId)
						if err != nil {
							return
						}
					}
				}
			}
		}
	}
}

// Checks if the assetid corresponds to a glutz device and that the value written to open is 1
func checkThereIsADoorToBeOpened(output api.Data) (bool, error) {
	DeviceExists, err := conf.ExistGlutzDeviceWithAssetId(context.Background(), output.AssetId)
	if err != nil {
		log.Error("Output", "Error checking if asset id corresponds to a glutz device")
		return false, err
	}
	data, err := mapToStruct(output.Data)
	if err != nil {
		log.Error("Output", "Error converting map to struct")
		return false, err
	}
	open := data.Open
	doorAlreadyOpen, err := checkDoorIfDoorIsAlreadyOpen(output.AssetId)
	if err != nil {
		log.Error("Output", "Error checking whether door is already open")
		return false, err
	}
	if !DeviceExists || open != 1 || doorAlreadyOpen {
		return false, nil
	}
	return true, nil
}

// Checks if a door is opened by reading the "openable" attribute for the asset with the given assetid.
func checkDoorIfDoorIsAlreadyOpen(assetid int32) (bool, error) {
	request, err := http.NewRequestWithApiKey(common.Getenv("API_ENDPOINT", "")+"/data?assetId="+strconv.Itoa(int(assetid))+"&dataSubtype=input", "X-API-KEY", common.Getenv("API_TOKEN", ""))
	if err != nil {
		log.Error("Output", "Error with request: %v", err)
		return false, err
	}
	asset_data, err := http.Read[glutz.AssetData](request, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("Output", "Error reading asset data: %v", err)
		return false, err
	}
	if asset_data[0].Data.Openable == 1 {
		log.Debug("Output", "Door is already open")
		return true, nil
	} else {
		return false, nil
	}

}

// Fetches the Glutz device where a value was changed in the database and the configuration
func getDeviceAndGetConfig(output api.Data) (*apiserver.Device, *apiserver.Configuration, error) {
	device, err := conf.GetDevicewithAssetId(context.Background(), output.AssetId)
	if err != nil {
		log.Error("Output", "Error getting device from assetid %v", err)
		return nil, nil, err
	}
	config, err := conf.GetConfig(context.Background(), int64(device.ConfigId))
	if err != nil {
		log.Error("Output", "Error getting configuration %v", err)
		return nil, nil, err
	}
	return device, config, nil
}

// Check if a value exists in glutz environment for openable duration for this door. If so, use this value.
// If not, use the default value from the config table
func getOpenableDuration(config *apiserver.Configuration, device *apiserver.Device) (int, error) {
	glutzOpenableDuration, err := getAccessPointPropertyOpenableDuration(*config, device.LocationId)
	if err != nil {
		log.Error("Output", "Error sending openable duration to door")
		return 0, err
	}
	var openableDuration int
	if glutzOpenableDuration != "" {
		openableDuration, err = strconv.Atoi(glutzOpenableDuration)
		if err != nil {
			log.Error("Output", "Couldn't convert to integer %v", err)
			return 0, err
		}

	} else {
		openableDuration = int(config.DefaultOpenableDuration)
	}
	return openableDuration, nil
}

// Opens/closes the door. Openable Duration isn't considered in the current Glutz API implementation
func sendOpenableDurationToDoor(config apiserver.Configuration, openableDuration int, locationid string) (bool, error) {
	durationstring := formatDuration(openableDuration)
	req := Request{
		Jsonrpc: "2.0",
		ID:      "m",
		Method:  "eAccess.openAccessPoint",
		Params: []interface{}{
			locationid,
			Duration{Duration: durationstring},
		},
	}
	setdurationrequest, err := http.NewPostRequest(config.Url+"/rpc", req)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
		return false, err
	}
	setdurationrequest.Header.Add("Referer", config.Url)
	setdurationrequest.SetBasicAuth(config.Username, config.Password)
	durationset, err := http.Read[glutz.Properties](setdurationrequest, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading device status: %v", err)
		return false, err
	}
	return durationset.Result, nil
}

// Waits until the time is ready to close door again. Then closes door.
func waitAndResetOpen(config apiserver.Configuration, openableDuration int, assetid int32, locationid string) {
	time.Sleep(time.Second * time.Duration(openableDuration))
	// Here we close the door again automatically after the length of time "openable duration" as it seems
	// the Glutz API doesn't take the time into account.
	response, _ := sendOpenableDurationToDoor(config, 0, locationid)
	if response {
		eliona.UpsertOpenData(0, assetid)
		log.Debug("Output", "Closed door at Location %v again", locationid)

	} else {
		eliona.UpsertOpenData(2, assetid)
	}
}

func mapToStruct(m map[string]interface{}) (*OutputData, error) {
	s := &OutputData{}

	if v, ok := m["open"].(float64); ok {
		s.Open = v
	} else {
		return nil, fmt.Errorf("invalid type for field 'openable'")
	}

	return s, nil
}

func formatDuration(duration int) string {
	hours := duration / 3600
	minutes := (duration % 3600) / 60
	seconds := duration % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func listenApiRequests() {
	err := nethttp.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"), utilshttp.NewCORSEnabledHandler(
		apiserver.NewRouter(
			apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
			apiserver.NewVersionApiController(apiservices.NewVersionApiService()),
			apiserver.NewCustomizationApiController(apiservices.NewCustomizationApiService()),
			apiserver.NewDevicesApiController(apiservices.NewDevicesApiService()),
		)))
	log.Fatal("main", "Error in API Server: %v", err)
}
