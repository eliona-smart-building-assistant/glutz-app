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
	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"glutz/conf"
	"glutz/eliona"
	"time"
	"context"
)

// The main function starts the app by starting all services necessary for this app and waits
// until all services are finished.
func main() {
	log.Info("main", "Starting the app.")

	// Necessary to close used init resources, because db.Pool() is used in this app.
	defer db.ClosePool()

	// Init the app before the first run.
	app.Init(db.Pool(), app.AppName(),
		asset.InitAssetTypeFile("eliona/asset-type-glutz_device.json"),
		app.ExecSqlFile("conf/init.sql"),
		conf.InitConfiguration,
		eliona.InitEliona,
	)

	var functions []func()
	functions = append(functions, listenApi)
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Error("Configurations", "Error retrieving configurations")
	}
	for _, config := range configs {
		log.Debug("main", "Appending processDevices() with configID: %v and refresh interval %v", config.ConfigId, config.RefreshInterval)
		functions = append(functions, common.LoopWithParam(processDevices, config.ConfigId, time.Duration(config.RefreshInterval)*time.Second))	
	}
	common.WaitFor(functions...)

	log.Info("main", "Terminate the app.")
}
