package client_test

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"

	"github.com/alqmy/weather-mesh/internal/pkg/messages"

	"github.com/alqmy/weather-mesh/internal/pkg/client"
	"github.com/pebbe/zmq4"
)

func TestPushWeatherUpdates(t *testing.T) {
	push, _ := zmq4.NewSocket(zmq4.PUSH)
	pull, _ := zmq4.NewSocket(zmq4.PULL)

	pull.Bind("tcp://*:6000")
	push.Connect("tcp://localhost:6000")

	defer pull.Close()
	defer push.Close()

	updates := make(chan messages.WeatherUpdate)
	defer close(updates)

	wg := sync.WaitGroup{}

	go func() {
		if err := client.PushWeatherUpdates(updates, push); err != nil {
			t.Fatal(err)
		}
	}()

	updates <- messages.WeatherUpdate{
		Temperature: 1.0,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		d, _ := json.Marshal(messages.WeatherUpdate{
			Temperature: 1.0,
		})
		data, _ := json.Marshal(messages.MessageWrapper{
			Type: "weather-update",
			Data: d,
		})

		raw, err := pull.RecvBytes(0)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !bytes.Equal(raw, data) {
			t.Fatalf("Expected %v got %v", data, raw)
		}
	}()

	wg.Wait()
}
