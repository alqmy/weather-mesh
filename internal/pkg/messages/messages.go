package messages

import (
	"encoding/json"

	"github.com/alqmy/weather-mesh/internal/pkg/members"
)

// MessageWrapper wraps a message
type MessageWrapper struct {
	Type string
	Data json.RawMessage
}

type ParticipationRequest struct {
	Address string
	Name    string
}

type ParticipationResponse struct {
	Valid bool
}

type ListRequest struct {
}

type ListResponse struct {
	Members []members.Member
}

// WeatherUpdate represents a weather update
type WeatherUpdate struct {
	Temperature float64 // Temperature in Centigrade
	Dewpoint    float64
	Humidity    float64 // Humidity in percentage
	Pressure    float64 // Pressure is in Pa
	Latitude    float64 // Latitude in degrees
	Longitude   float64 // Longitude in degrees
	NodeName    string
}
