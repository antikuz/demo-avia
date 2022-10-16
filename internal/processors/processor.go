package processors

import (
	"math/rand"
	"net/url"
	"time"

	"github.com/antikuz/demo-avia/internal/db"
	"github.com/antikuz/demo-avia/internal/models"
	"github.com/antikuz/demo-avia/pkg/logging"
)

type StorageProcessor struct {
	storage *db.Storage
	logger *logging.Logger
}

func NewStorageProcessor(storage *db.Storage, logger *logging.Logger) *StorageProcessor {
	return &StorageProcessor{
		storage: storage,
	}
}

func (s *StorageProcessor) List(postFormValues url.Values) []models.FlightsV {
	departure_city := postFormValues["departure"][0]
	arrival_city := postFormValues["arrival"][0]
	dateFromString := postFormValues["date"][0]
	passengersCount := postFormValues["passengers_count"][0]
	class := postFormValues["class"][0]

	dateFrom, err := time.Parse("2006-01-02", dateFromString)
	if err != nil {
		s.logger.Errorf("Failed to parse date string, due to err: %v\n", err)
	}
	dateTo := dateFrom.AddDate(0,0,1)
	
	return s.storage.List(departure_city, arrival_city, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"), passengersCount, class)
}

func (s *StorageProcessor) GetFlight(flightID string) models.FlightsV {
	return s.storage.GetFlight(flightID)
}

func (s *StorageProcessor) GetUser(username string) models.User {
	users := s.storage.GetUser(username)
	if len(users) != 1 {
		return models.User{}
	}
	return users[0]
}

func (s *StorageProcessor) GetUserFlights(username string) []models.UserFlights {
	return s.storage.GetUserFlights(username)
}

func (s *StorageProcessor) BuyTicket(formValues url.Values) bool {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	book_ref := make([]rune, 6)
	for i := range book_ref {
		book_ref[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	ticket_no := make([]rune, 13)
	for i := range ticket_no {
		ticket_no[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	flightid := formValues["flight_id"][0]
	passenger_name := formValues["name"][0]
	passenger_id := formValues["passport"][0]
	fare_conditions := formValues["class"][0]

	err := s.storage.BuyTicket(string(book_ref), string(ticket_no), passenger_id, passenger_name, fare_conditions, flightid)
	return err == nil
}