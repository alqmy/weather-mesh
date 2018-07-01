package main

import (
	"log"

	"github.com/pebbe/zmq4"

	"github.com/alqmy/weather-mesh/internal/pkg/client"
	"github.com/alqmy/weather-mesh/internal/pkg/messages"
)

var (
	weatherFile     = "/var/www/html/Output.json"
	connectionPoint = "tcp://192.168.1.64:3132"
)

func main() {
	c, _ := zmq4.NewContext()
	push, _ := c.NewSocket(zmq4.PUSH)

	if err := push.Connect(connectionPoint); err != nil {
		log.Fatal(err)
	}

	name := client.GenerateName()

	errChan := make(chan error)
	updates := make(chan messages.WeatherUpdate)
	defer close(updates)

	go func() {
		for {
			update, err := client.ReadWeatherUpdate(weatherFile)
			if err != nil {
				errChan <- err
			}

			updates <- messages.WeatherUpdate{
				Temperature: update.Temperature,
				Humidity:    update.Humidity,
				Pressure:    update.Pressure,
				// Latitude:    update.Location.Latitude,
				// Longitude:   update.Location.Longitude,
				Dewpoint: update.Dewpoint,
				NodeName: name,
			}
		}
	}()

	go func() {
		if err := client.PushWeatherUpdates(updates, push); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		log.Fatal(err)
	}
}
