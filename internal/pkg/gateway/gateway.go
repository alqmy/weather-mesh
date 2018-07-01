package gateway

import (
	"context"
	"encoding/json"
	"log"

	"github.com/alqmy/weather-mesh/internal/pkg/messages"
	"github.com/pebbe/zmq4"
)

// PublishWeatherUpdates publishes a stream of weather updates to a zmq Pub socket
func PublishWeatherUpdates(updates <-chan messages.WeatherUpdate, pub *zmq4.Socket) error {

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

		_, err = pub.SendBytes(raw, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

// PullWeatherUpdates listens on a pull socket and pushes updates to a channel
func PullWeatherUpdates(ctx context.Context, pull *zmq4.Socket, updates chan<- messages.WeatherUpdate) error {

	for {
		raw, err := pull.RecvBytes(0)
		if err != nil {
			return err
		}
		message := new(messages.MessageWrapper)

		err = json.Unmarshal(raw, message)
		if err != nil {
			log.Printf("%##v\n", err)
			continue
		}

		switch message.Type {
		case "weather-update":
			update := messages.WeatherUpdate{}

			err = json.Unmarshal(message.Data, &update)
			if err != nil {
				log.Printf("%##v\n", err)
				continue
			}

			updates <- update
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
}
