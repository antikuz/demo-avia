package db

import (
	"context"
	"fmt"

	"github.com/antikuz/demo-avia/internal/models"
	"github.com/antikuz/demo-avia/pkg/logging"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Storage struct {
	databasePool *pgxpool.Pool
	logger *logging.Logger
}

func NewStorage(pool *pgxpool.Pool, logger *logging.Logger) *Storage {
	return &Storage{
		databasePool: pool,
		logger: logger,
	}
}

func (storage *Storage) GetFlight(flightID string) models.FlightsV  {
	query := fmt.Sprintf(`
	SELECT flight_id, 
	       scheduled_departure_local, 
		   scheduled_arrival_local, 
		   scheduled_duration, 
		   departure_city, 
		   arrival_city 
	FROM flights_v fv
    WHERE flight_id = '%s'`, flightID)
	var result []models.FlightsV
	err := pgxscan.Select(context.Background(), storage.databasePool, &result, query)
	if err != nil {
		storage.logger.Errorf("Failed to list metrics, due to err: %v\n", err)
	}
	return result[0]
}



func (storage *Storage) List(departure_city string, arrival_city string, dateFrom string, dateTo string) []models.FlightsV {
	query := fmt.Sprintf(`
	SELECT flight_id, 
	       scheduled_departure_local, 
		   scheduled_arrival_local, 
		   scheduled_duration, 
		   departure_city, 
		   arrival_city 
	FROM flights_v fv
	WHERE departure_city = '%s'
	  AND arrival_city = '%s'
	  AND scheduled_departure_local >= '%s'
	  AND scheduled_departure_local < '%s'
	  AND (SELECT COUNT(*) - (SELECT COUNT(*)
		 FROM ticket_flights tf
		 WHERE tf.flight_id = fv.flight_id
		  AND tf.fare_conditions = 'Business')
	FROM seats
	WHERE aircraft_code = fv.aircraft_code 
	AND fare_conditions = 'Business') > 0;`,departure_city, arrival_city, dateFrom, dateTo)
	var result []models.FlightsV
	err := pgxscan.Select(context.Background(), storage.databasePool, &result, query)
	if err != nil {
		storage.logger.Errorf("Failed to list metrics, due to err: %v\n", err)
	}

	return result
}