package client

import (
	"encoding/json"
	"math/rand"
	"os"

	"github.com/alqmy/weather-mesh/internal/pkg/messages"
	"github.com/pebbe/zmq4"
)

type WeatherSnapshot struct {
	Temperature float64 `json:"temperature"`
	Dewpoint    float64 `json:"dewpoint"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
	Location    struct {
		Longitude float64 `json:"lon"`
		Latitude  float64 `json:"lat"`
	} `json:"location"`
	// At time.Time `json:"timecollected"`
}

func GenerateName() string {
	n := rand.Intn(len(Nouns))
	a := rand.Intn(len(Adjectives))

	noun := Nouns[n]
	adjective := Adjectives[a]

	return adjective + " " + noun
}

func PushWeatherUpdates(updates <-chan messages.WeatherUpdate, push *zmq4.Socket) error {

	for update := range updates {
		data, err := json.Marshal(update)
		if err != nil {
			return err
		}

		message := messages.MessageWrapper{
			Type: "weather-update",
			Data: data,
		}

		raw, err := json.Marshal(message)
		if err != nil {
			return err
		}

		_, err = push.SendBytes(raw, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadWeatherUpdate(filename string) (*WeatherSnapshot, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(file)

	update := WeatherSnapshot{}
	err = d.Decode(&update)
	if err != nil {
		return nil, err
	}

	return &update, nil
}
