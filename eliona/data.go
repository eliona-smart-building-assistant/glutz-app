package eliona

import (
	"fmt"
	"glutz/glutz"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

type deviceInputDataPayload struct {
	BatteryLevel     int64 `json:"batteryLevel"`
	Openings         int64   `json:"openings"`
}

type deviceInfoDataPayload struct {
	Building         string  `json:"building"`
	Room             string  `json:"room"`
	AccessPoint      string   `json:"accessPoint"`
	OperatingMode    int64   `json:"operatingMode"`
	Firmware         string  `json:"firmware"`
}

func UpsertInputData(deviceData glutz.DeviceDb, assetId int32) error{
	log.Debug("Data", "Uploading input data")
	deviceInput:= deviceInputDataPayload{
		BatteryLevel: deviceData.BatteryLevel,
		Openings: deviceData.Openings,
	}
	err:= upsertData(api.SUBTYPE_INPUT, assetId, deviceInput)
	if err!= nil {
		log.Error("Data", "Error sending input data")
		return err
	}
	return nil
}

func UpsertInfoData(deviceData glutz.DeviceDb, assetId int32) error{
	log.Debug("Data", "Uploading info data")
	deviceInfo:= deviceInfoDataPayload{
		Building: deviceData.Building,
		Room: deviceData.Room,
		AccessPoint: deviceData.AccessPoint,
		OperatingMode: deviceData.OperatingMode,
		Firmware: deviceData.Firmware,
	}
	err:= upsertData(api.SUBTYPE_INFO, assetId, deviceInfo)
	if err!= nil {
		log.Error("Data", "Error sending info data")
		return err
	}
	return nil
	

}

func upsertData(subtype api.DataSubtype, assetId int32, payload any) error {
	var statusData api.Data
	statusData.Subtype = subtype
	now := time.Now()
	statusData.Timestamp = *api.NewNullableTime(&now)
	statusData.AssetId = assetId
	statusData.Data = common.StructToMap(payload)
	if err := asset.UpsertDataIfAssetExists[any](statusData); err != nil {
		return fmt.Errorf("upserting data: %v", err)
	}
	return nil
}

