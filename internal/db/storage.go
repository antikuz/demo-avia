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
		storage.logger.Errorf("Failed to get flight, due to err: %v\n", err)
	}
	return result[0]
}

func (storage *Storage) GetUser(username string) []models.User  {
	query := fmt.Sprintf(`
	SELECT *
	FROM users
    WHERE username = '%s'`, username)
	var result []models.User
	err := pgxscan.Select(context.Background(), storage.databasePool, &result, query)
	if err != nil {
		storage.logger.Errorf("Failed to get users, due to err: %v\n", err)
	}
	return result
}

func (storage *Storage) GetUserFlights(username string) []models.UserFlights {
	query := fmt.Sprintf(`
	SELECT	f.scheduled_departure,
			f.departure_city || ' (' || f.departure_airport || ')' AS departure,
			f.arrival_city || ' (' || f.arrival_airport || ')' AS arrival,
			tf.flight_id
	FROM ticket_flights tf
	JOIN tickets t ON t.ticket_no = tf.ticket_no
	JOIN users u ON t.passenger_id = u.passenger_id
	JOIN flights_v f ON tf.flight_id = f.flight_id
	WHERE    u.username = '%s'
	ORDER BY f.scheduled_departure;`, username)
	var result []models.UserFlights
	storage.logger.Debugf("%+v", result)
	err := pgxscan.Select(context.Background(), storage.databasePool, &result, query)
	if err != nil {
		storage.logger.Errorf("Failed to get users, due to err: %v\n", err)
	}
	return result
}

func (storage *Storage) List(departure_city string, arrival_city string, dateFrom string, dateTo string, passengersCount string, class string) []models.FlightsV {
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
		  AND tf.fare_conditions = '%s')
	FROM seats
	WHERE aircraft_code = fv.aircraft_code 
	AND fare_conditions = 'Business') >= %s`,departure_city, arrival_city, dateFrom, dateTo, class, passengersCount)
	var result []models.FlightsV
	err := pgxscan.Select(context.Background(), storage.databasePool, &result, query)
	if err != nil {
		storage.logger.Errorf("Failed to list list flights, due to err: %v\n", err)
	}

	return result
}