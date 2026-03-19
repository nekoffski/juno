package device

type DeviceVendor string
type DeviceStatus string

const (
	DeviceVendorYeelight DeviceVendor = "Yeelight"
	DeviceVendorXiaomi   DeviceVendor = "Xiaomi"
	DeviceStatusOnline   DeviceStatus = "online"
	DeviceStatusOffline  DeviceStatus = "offline"
	DeviceStatusIdle     DeviceStatus = "idle"
)

type DeviceModel struct {
	Id           int          `json:"id"`
	Name         string       `json:"name"`
	Vendor       DeviceVendor `json:"vendor"`
	Status       DeviceStatus `json:"status"`
	Capabilities []string     `json:"capabilities"`
}

type DeviceAddr struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type Device interface {
	Model() *DeviceModel
}
