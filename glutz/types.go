package glutz



type DevicesDb map[string]DeviceDb

type DeviceDb struct {
	BatteryLevel     int64 `json:"batteryLevel"`
	Openings         int64   `json:"openings"`
	Building         string  `json:"building"`
	Room             string  `json:"room"`
	AccessPoint      string   `json:"accessPoint"`
	OperatingMode    int64   `json:"operatingMode"`
	Firmware         string  `json:"firmware"`
	Openable 		 bool  `json:"openable"` 
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

type DeviceStatusGlutz struct{
	Id 			string `json:"id"`
	Jsonrpc		string `json:"jsonrpc"`
	Result 		[]DeviceStatus `json:"result"`
}

type DeviceStatus struct{
	BatteryAlarm	bool `json:"batteryAlarm"`
	BatteryLevel	int64 `json:"batteryLevel"`
	BatteryPowered	bool `json:"batteryPowered"`
	CommunicationErrors	int64 `json:"communicationErrors"`
	DeviceType	int64 `json:"deviceType"`
	DeviceId	string `json:"deviceid"`
	Firmware	string `json:"firmware"`
	IrWakeups	int64 `json:"irWakeups"`
	LastError	int64  `json:"lastError"`
	LastUpdate	string `json:"lastUpdate"`
	Openings	int64 `json:"openings"`
	OperatingMode	int64 `json:"operatingMode"`
	RfWakeups	int64 `json:"rfWakeups"`
}

type DeviceAccessPointGlutz struct{
	Id 			string `json:"id"`
	Jsonrpc		string `json:"jsonrpc"`
	Result 		[]string `json:"result"`
}

type Properties struct{
	Id 			string `json:"id"`
	Jsonrpc		string `json:"jsonrpc"`
	Result 		bool `json:"result"`
}

type GlutzOpenableDuration struct{
	Id 			string `json:"id"`
	Jsonrpc		string `json:"jsonrpc"`
	Result 		string `json:"result"`
}