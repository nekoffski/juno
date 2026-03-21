package device

type DeviceVendor string
type DeviceStatus string

const (
	DeviceVendorYeelight DeviceVendor = "Yeelight"
	DeviceVendorXiaomi   DeviceVendor = "Xiaomi"
)

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusIdle    DeviceStatus = "idle"
)

type ColorRGB struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
}

type DeviceAddr struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type DeviceModel struct {
	Id           int          `json:"id"`
	Name         string       `json:"name"`
	Vendor       DeviceVendor `json:"vendor"`
	Status       DeviceStatus `json:"status"`
	Addr         DeviceAddr   `json:"addr"`
	Capabilities []string     `json:"capabilities"`
}

type Action struct {
	Method string `json:"method"`
	Params any    `json:"params"`
}

type Device interface {
	Model() *DeviceModel
	Properties() map[string]any
	IsCapable(action string) bool
	EnqueueAction(action Action) error
}
