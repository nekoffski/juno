package device

type GetDevicesRequest struct{}

type GetDevicesResponse struct {
	Devices []*DeviceModel
}

type GetDeviceByIdRequest struct {
	Id int `json:"id"`
}

type GetDeviceByIdResponse struct {
	Device *DeviceModel `json:"device"`
}

type DiscoverDevicesRequest struct{}

type AckResponse struct{}
