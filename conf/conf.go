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

package conf

import (
	"context"
	"glutz/apiserver"
	dbglutz "glutz/db/glutz"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func GetDevices(ctx context.Context, configId int64) ([]apiserver.Device, error) {
	var mods []qm.QueryMod 
	if configId > 0 {
		mods = append(mods, dbglutz.DeviceWhere.ConfigID.EQ(configId))
	}
	dbSpaces, err := dbglutz.Devices(mods...).All(ctx, db.Database("glutz"))
	if err != nil {
		return nil, err
	}
	var apiDevices []apiserver.Device
	for _, dbDevices := range dbSpaces {
		apiDevices = append(apiDevices, *apiDevicesFromDbDevices(dbDevices))
	}
	return apiDevices, nil
}

func GetDevice(ctx context.Context, configId int64, projectId string, deviceId string) (*apiserver.Device, error) {
	var mods []qm.QueryMod 
	if configId > 0 {
		mods = append(mods, dbglutz.DeviceWhere.ConfigID.EQ(configId))
		mods = append(mods, dbglutz.DeviceWhere.ProjectID.EQ(projectId))
		mods = append(mods, dbglutz.DeviceWhere.DeviceID.EQ(deviceId))
	}
	dbDevices, err := dbglutz.Devices(mods...).All(ctx, db.Database("glutz")) 
	if err != nil {
		return nil, err
	}
	if len(dbDevices)!= 1 {
		return nil, nil
	}
	return apiDevicesFromDbDevices(dbDevices[0]), nil
}


func DeleteConfig(ctx context.Context, configId int64) (int64, error) {
	return dbglutz.Configs(dbglutz.ConfigWhere.ConfigID.EQ(configId)).DeleteAll(ctx, db.Database("glutz"))
}

func GetConfig(ctx context.Context, configId int64) (*apiserver.Configuration, error) {
	dbConfigs, err := dbglutz.Configs(dbglutz.ConfigWhere.ConfigID.EQ(configId)).All(ctx, db.Database("glutz"))
	if err != nil {
		return nil, err
	}
	if len(dbConfigs) == 0 {
		return nil, err
	}
	return apiConfigFromDbConfig(dbConfigs[0]), nil
}

func GetConfigs(ctx context.Context) ([]apiserver.Configuration, error) {
	dbConfigs, err := dbglutz.Configs().All(ctx, db.Database("glutz"))
	if err != nil {
		return nil, err
	}
	var apiConfigs []apiserver.Configuration
	for _, dbConfig := range dbConfigs {
		apiConfigs = append(apiConfigs, *apiConfigFromDbConfig(dbConfig))
	}
	return apiConfigs, nil
}

func InsertConfig(ctx context.Context, config apiserver.Configuration) (apiserver.Configuration, error) {
	dbConfig := dbConfigFromApiConfig(&config)
	err := dbConfig.Insert(ctx, db.Database("glutz"), boil.Blacklist(dbglutz.ConfigColumns.ConfigID))
	if err != nil {
		return apiserver.Configuration{}, err
	}
	config.ConfigId = dbConfig.ConfigID
	return config, err
}

func UpsertConfigById(ctx context.Context, configId int64, config apiserver.Configuration) (apiserver.Configuration, error) {
	dbConfig := dbConfigFromApiConfig(&config)
	dbConfig.ConfigID = configId
	err := dbConfig.Upsert(ctx, db.Database("glutz"), true,
		[]string{dbglutz.ConfigColumns.ConfigID},
		boil.Blacklist(dbglutz.ConfigColumns.ConfigID),
		boil.Infer(),
	)
	config.ConfigId = dbConfig.ConfigID
	return config, err
}

func InsertSpace(ctx context.Context, configId int64, projectId string, deviceId string, assetId int32, locationId string) error {
	var dbDevice dbglutz.Device
	dbDevice.ConfigID = configId
	dbDevice.ProjectID = projectId
	dbDevice.DeviceID = deviceId
	dbDevice.AssetID = assetId
	dbDevice.LocationID = locationId
	return dbDevice.Insert(ctx, db.Database("glutz"), boil.Infer())
}

func SetConfigActiveState(configID int64, state bool) (int64, error) {
	return dbglutz.Configs(
		dbglutz.ConfigWhere.ConfigID.EQ(null.Int64FromPtr(&configID).Int64),
	).UpdateAll(context.Background(), db.Database("glutz"), dbglutz.M{
		dbglutz.ConfigColumns.Active: state,
	})
}

func IsConfigActive(config apiserver.Configuration) bool {
	return config.Active == nil || *config.Active
}

func IsConfigEnabled(config apiserver.Configuration) bool {
	return config.Enable == nil || *config.Enable
}

func SetAllConfigsInactive(ctx context.Context) (int64, error) {
	return dbglutz.Configs().UpdateAllG(ctx, dbglutz.M{
		dbglutz.ConfigColumns.Active: false,
	})
}



///// API to DB Mappings //////

func apiDevicesFromDbDevices(dbDevices *dbglutz.Device) *apiserver.Device {
	var apiDevices apiserver.Device
	apiDevices.ConfigId = int32(dbDevices.ConfigID)
	apiDevices.ProjectId = dbDevices.ProjectID
	apiDevices.AssetId = dbDevices.AssetID
	apiDevices.LocationId = dbDevices.LocationID
	return &apiDevices
}

func apiConfigFromDbConfig(dbConfig *dbglutz.Config) *apiserver.Configuration {
	var apiConfig apiserver.Configuration
	apiConfig.ConfigId = dbConfig.ConfigID
	apiConfig.Username = dbConfig.Username
	apiConfig.Password = dbConfig.Password
	apiConfig.ApiToken = dbConfig.APIToken
	apiConfig.Url = dbConfig.URL
	apiConfig.Active = &dbConfig.Active.Bool
	apiConfig.Enable = &dbConfig.Enable.Bool
	apiConfig.RequestTimeout = dbConfig.RequestTimeout.Int32
	apiConfig.RefreshInterval = dbConfig.RefreshInterval.Int32
	apiConfig.ProjIds = common.Ptr[[]string](dbConfig.ProjectIds)
	return &apiConfig
}

func dbConfigFromApiConfig(apiConfig *apiserver.Configuration) *dbglutz.Config {
	var dbConfig dbglutz.Config
	dbConfig.ConfigID = null.Int64FromPtr(&apiConfig.ConfigId).Int64
	dbConfig.Username = apiConfig.Username
	dbConfig.Password = apiConfig.Password
	dbConfig.APIToken = apiConfig.ApiToken
	dbConfig.URL = apiConfig.Url
	dbConfig.Active = null.BoolFromPtr(apiConfig.Active)
	dbConfig.Enable = null.BoolFromPtr(apiConfig.Enable)
	dbConfig.RefreshInterval = null.Int32FromPtr(&apiConfig.RefreshInterval)
	dbConfig.RequestTimeout = null.Int32FromPtr(&apiConfig.RequestTimeout)
	if apiConfig.ProjIds != nil {
		dbConfig.ProjectIds = *apiConfig.ProjIds
	}
	return &dbConfig
}

