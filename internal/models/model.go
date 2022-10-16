package models

import (
	"net/url"
	"time"
)

type SearchResult struct {
	SearchValues url.Values
	SearchResults []FlightsV
}

type BuyFlightID struct {
	SearchValues url.Values
	SearchResults FlightsV
}

type FlightsV struct {
	FlightID int `db:"flight_id"`
	DepartureCity  string `db:"departure_city"`
	ArrivalCity string `db:"arrival_city"`
	DepartureDate time.Time `db:"scheduled_departure_local"`
	ArrivalDate time.Time `db:"scheduled_arrival_local"`
	FlyDuration time.Duration `db:"scheduled_duration"`
}

