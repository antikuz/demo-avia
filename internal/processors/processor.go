package processors

import (
	"time"
	"net/url"

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

	dateFrom, err := time.Parse("2006-01-02", dateFromString)
	if err != nil {
		s.logger.Errorf("Failed to parse date string, due to err: %v\n", err)
	}
	dateTo := dateFrom.AddDate(0,0,1)
	
	return s.storage.List(departure_city, arrival_city, dateFrom.Format("2006-01-02"), dateTo.Format("2006-01-02"))
}

func (s *StorageProcessor) GetFlight(FlightID string) models.FlightsV {
	return s.storage.GetFlight(FlightID)
}