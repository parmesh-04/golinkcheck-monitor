// api/types.go

package api

// CreateMonitorRequest defines the shape of the JSON body for creating a monitor.
type CreateMonitorRequest struct {
	URL         string `json:"url" validate:"required,url"`
	IntervalSec int    `json:"intervalSec" validate:"required,gt=0,max=86400"` // Max 1 day
}

// UpdateMonitorRequest defines the shape of the JSON body for updating a monitor.
type UpdateMonitorRequest struct {
	URL         string `json:"url" validate:"required,url"`
	IntervalSec int    `json:"intervalSec" validate:"required,gt=0,max=86400"`
	Active      bool   `json:"active"` // 'active' is optional, so no 'required' tag
}