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

type deviceRequest struct {
	jsonrpc string
	id      string
	method  string
	params  []string
}

// doAnything is the main app function which is called periodically
func processDevices(configId int64) {
	config, err := conf.GetConfig(context.Background(), configId)
	if err != nil {
		log.Error("devices", "Error reading configuration: %v", err)
	}
	var body deviceRequest
	strArr := [1]string{"Devices"}
	body.id = "m"
	body.jsonrpc = "2.0"
	body.method = "eAccess.getModel"
	body.params = strArr[:]
	log.Debug("devices", "Request Body: %v", body)
	request, err := http.NewPostRequest(config.Url, body)
	if err != nil {
		log.Error("devices", "Error with request: %v", err)
	}
	deviceList, err := http.Read[glutz.DeviceGlutz](request, time.Duration(time.Duration.Seconds(1)), true)
	if err != nil {
		log.Error("devices", "Error reading spaces: %v", err)
	}

	log.Debug("Devices", "Here are the devices: %v", deviceList)

	// for device:= range devices {
	// 	log.Debug("devices", "Device: %v", device)
	// }

}

// listenApi starts the API server and listen for requests
func listenApi() {
	http.ListenApiWithOs(&nethttp.Server{Addr: ":" + common.Getenv("API_SERVER_PORT", "3000"), Handler: apiserver.NewRouter(
		apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
		apiserver.NewVersionApiController(apiservices.NewVersionApiService()),
	)})
}
