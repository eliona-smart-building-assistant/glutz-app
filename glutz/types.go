package glutz



type DevicesDb map[string]DeviceDb

type DeviceDb struct {
	BatteryLevel     float64 `json:"batteryLevel"`
	Openings         int64   `json:"openings"`
	Building         string  `json:"building"`
	Room             string  `json:"room"`
	AccessPoint      int64   `json:"accessPoint"`
	OperatingMode    int64   `json:"operatingMode"`
	Firmware         string  `json:"firmware"`
	OpenableDuration string  `json:"openableDuration"` // Check this!
}


type DeviceGlutz struct {
	Id 			string `json:"id"`
	Jsonrpc		string `json:"jsonrpc"`
	Result 		[]DeviceResult `json:"result"`
}

type DeviceResult struct {
	AccessPointId 	string `json:"accessPointId"`
	DeviceType 		int64 `json:"deviceType"`
	Deviceid 		string `json:"deviceid"`
	Id 				string  `json:"id"`
	Label 			string `json:"label"`
}