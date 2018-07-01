package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/alqmy/weather-mesh/internal/pkg/client"

	"github.com/pebbe/zmq4"

	"database/sql"

	"github.com/alqmy/weather-mesh/internal/pkg/gateway"
	m "github.com/alqmy/weather-mesh/internal/pkg/members"
	"github.com/alqmy/weather-mesh/internal/pkg/messages"
	_ "github.com/mattn/go-sqlite3"
)

var (
	memMux  = sync.Mutex{}
	members m.Members
)

func duplicateStreams(updates <-chan messages.WeatherUpdate, streams ...chan<- messages.WeatherUpdate) {
	for update := range updates {
		for _, stream := range streams {
			stream <- update
		}
	}
}

func main() {
	c, err := zmq4.NewContext()
	if err != nil {
		log.Fatal(err)
	}

	pull, err := c.NewSocket(zmq4.PULL)
	if err != nil {
		log.Fatal(err)
	}
	if err := pull.Bind("tcp://*:3132"); err != nil {
		log.Fatal(err)
	}
	defer pull.Close()

	pub, err := c.NewSocket(zmq4.PUB)
	if err != nil {
		log.Fatal(err)
	}
	if err := pub.Bind("tcp://*:5000"); err != nil {
		log.Fatal(err)
	}
	defer pub.Close()

	sub, err := c.NewSocket(zmq4.SUB)
	if err != nil {
		log.Fatal(err)
	}
	if err := sub.Connect("tcp://localhost:5000"); err != nil {
		log.Fatal(err)
	}
	defer sub.Close()
	sub.SetSubscribe("")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Exec(`CREATE TABLE updates (
		temperature REAL,
		humidity REAL,
		pressure REAL
	)`)

	db.Exec(`INSERT INTO updates (temperature, humidity, pressure) VALUES (0,0,0)`)

	errChan := make(chan error, 1)

	// main process context for graceful cancellation
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	updates := make(chan messages.WeatherUpdate, 100)
	defer close(updates)

	// Start pulling weather updates
	go func() {
		errChan <- gateway.PullWeatherUpdates(ctx, pull, updates, func(update messages.WeatherUpdate) {
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}

			_, err = tx.Exec(
				"INSERT INTO updates (temperature, humidity, pressure) VALUES (?,?,?)",
				update.Temperature,
				update.Humidity,
				update.Pressure,
			)
			if err != nil {
				tx.Rollback()
				errChan <- err
				return
			}

			tx.Commit()
		})
	}()

	go func() {
		errChan <- gateway.PublishWeatherUpdates(updates, pub)
	}()

	h := http.NewServeMux()
	h.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		row := db.QueryRow("SELECT temperature, humidity, pressure FROM updates")

		snap := client.WeatherSnapshot{}

		err := row.Scan(
			&snap.Temperature,
			&snap.Humidity,
			&snap.Pressure,
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snap)
	})

	go func() {
		errChan <- http.ListenAndServe(":8080", h)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			log.Fatal(err)
		}
	}
}
