package glutz


type Devices map[string]Device


type Device struct {
	BatteryLevel float64 `json:"batteryLevel"`
	Openings int64 `json:"openings"`
	Building string `json:"building"`
	Room string `json:"room"`
	AccessPoint int64 `json:"accessPoint"`
	OperatingMode int64 `json:"operatingMode"`
	Firmware string `json:"firmware"`
	OpenableDuration string `json:"openableDuration"` // Check this!
}