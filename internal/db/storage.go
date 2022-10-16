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

func (storage *Storage) BuyTicket(book_ref string, ticket_no string, passenger_id string, passenger_name string, fare_conditions string, flightID string) error {
	queryBookings := fmt.Sprintf(`
	INSERT INTO bookings (book_ref, book_date, total_amount)
	VALUES      ('%s', bookings.now(), 0);`, book_ref)
	queryTickets := fmt.Sprintf(`
	INSERT INTO tickets (ticket_no, book_ref, passenger_id, passenger_name)
	VALUES      ('%s', '%s', '%s', '%s');`, ticket_no, book_ref, passenger_id, passenger_name)
	queryTicketFlights := fmt.Sprintf(`
	INSERT INTO ticket_flights (ticket_no, flight_id, fare_conditions, amount)
	VALUES      ('%s', '%s', '%s', 0);`, ticket_no, flightID, fare_conditions)
	storage.logger.Debug(queryBookings)
	storage.logger.Debug(queryTickets)
	storage.logger.Debug(queryTicketFlights)
	ctx := context.Background()
	transaction, err := storage.databasePool.Begin(ctx)
	if err != nil {
		storage.logger.Errorf("Begun transaction failed due to err: %v\n", err)
		return err
	}
	
	_, err = transaction.Exec(context.Background(), queryBookings)
	if err != nil {
		storage.logger.Error(err)
		err = transaction.Rollback(context.Background())
		if err != nil {
			storage.logger.Errorf("Rollback failed due to err: %v\n", err)
		}
		return err
	}

	_, err = transaction.Exec(context.Background(), queryTickets)
	if err != nil {
		storage.logger.Error(err)
		err = transaction.Rollback(context.Background())
		if err != nil {
			storage.logger.Errorf("Rollback failed due to err: %v\n", err)
		}
		return err
	}

	_, err = transaction.Exec(context.Background(), queryTicketFlights)
	if err != nil {
		storage.logger.Error(err)
		err = transaction.Rollback(context.Background())
		if err != nil {
			storage.logger.Errorf("Rollback failed due to err: %v\n", err)
		}
		return err
	}
	err = transaction.Commit(context.Background())
	if err != nil {
		storage.logger.Errorf("Transaction commit failed due to err: %v\n", err)
	}

	return err
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