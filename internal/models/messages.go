package models

type HeartbeatRequest struct {
	Magic string
}

type HeartbeatResponse struct {
	Healthy bool
	Magic   string
}
