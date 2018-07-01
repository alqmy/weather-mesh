package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/alqmy/weather-mesh/internal/pkg/gateway"
	"github.com/alqmy/weather-mesh/internal/pkg/messages"
	"github.com/pebbe/zmq4"
)

func TestPublishWeatherUpdates(t *testing.T) {
	pub, _ := zmq4.NewSocket(zmq4.PUB)
	sub, _ := zmq4.NewSocket(zmq4.SUB)

	pub.Bind("tcp://*:5000")
	<-time.After(2 * time.Second)
	sub.Connect("tcp://localhost:5000")
	sub.SetSubscribe("")

	defer pub.Close()
	defer pub.Close()

	updates := make(chan messages.WeatherUpdate)

	wg := sync.WaitGroup{}

	go func() {
		if err := gateway.PublishWeatherUpdates(updates, pub); err != nil {
			t.Fatal(err)
		}
	}()

	up := messages.WeatherUpdate{
		Temperature: 1.4,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		d, err := json.Marshal(up)
		raw, err := sub.RecvBytes(0)
		if err != nil {
			t.Fatal(err)
		}

		t.Log(raw)

		wrap := messages.MessageWrapper{}
		_ = json.Unmarshal(raw, &wrap)

		if wrap.Type != "weather-update" {
			t.Fatalf("Expected %s got %s", "weather-update", wrap.Type)
		}

		if !bytes.Equal(d, wrap.Data) {
			t.Fatalf("Got %s wanted %s", wrap.Data, d)
		}
	}()

	for i := 0; i < 100; i++ {
		updates <- up
	}
	close(updates)

	wg.Wait()
}

func TestPullWeatherUpdates(t *testing.T) {
	pull, _ := zmq4.NewSocket(zmq4.PULL)
	push, _ := zmq4.NewSocket(zmq4.PUSH)

	pull.Bind("tcp://*:6000")
	push.Connect("tcp://localhost:6000")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	updates := make(chan messages.WeatherUpdate)
	defer close(updates)

	go func() {
		if err := gateway.PullWeatherUpdates(ctx, pull, updates); err != nil {
			t.Fatal(err)
		}
	}()

	d, _ := json.Marshal(&messages.WeatherUpdate{
		Temperature: 5.0,
	})
	data, _ := json.Marshal(messages.MessageWrapper{
		Type: "weather-update",
		Data: d,
	})
	push.SendBytes(data, 0)

	select {
	case update := <-updates:
		if update.Temperature != 5.0 {
			t.Fail()
		}
	case <-time.After(5 * time.Second):
		t.Fail()
	}
}
