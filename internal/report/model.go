package report

type CreateReportRequest struct {
	EventType    string  `json:"event_type" binding:"required,oneof=armed_conflict use_of_force explosion terrorism civil_unrest other"`
	Severity     string  `json:"severity" binding:"required,oneof=critical high medium low"`
	Title        string  `json:"title" binding:"required,max=200"`
	Description  string  `json:"description" binding:"max=2000"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	LocationName string  `json:"location_name"`
	Country      string  `json:"country" binding:"required"`
}
