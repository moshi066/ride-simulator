package entities

type Ride struct {
	ID         string `json:"id"`
	DriverID   string `json:"driver_id"`
	RiderID    string `json:"rider_id"`
	RiderPhone string `json:"rider_phone"`
	Status     string `json:"status"`
}
type RideWithoutRiderPhone struct {
	ID       string `json:"id"`
	DriverID string `json:"driver_id"`
	RiderID  string `json:"rider_id"`
	Status   string `json:"status"`
}

type User struct {
	Phone  string `json:"phone"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type UserWithoutStatus struct {
	Phone string `json:"phone"`
	ID    string `json:"id"`
	Type  string `json:"type"`
}
