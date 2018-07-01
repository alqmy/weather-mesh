package client_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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

func TestReadWeatherUpdate(t *testing.T) {
	ioutil.WriteFile("./data.json", []byte(`{"temperature": 70.86, "dewpoint": 61.09, "humidity": 72.85, "pressure": 29.97, "location": {"lat": 0, "lon": 0}, "timecollected": "2018-07-01 03:20:44", "sensorID": "Node-01"}`), 0644)

	update, err := client.ReadWeatherUpdate("./data.json")
	if err != nil {
		t.Fatal(err)
	}

	if update == nil {
		t.Fatal("Expected not nil")
	}
}
