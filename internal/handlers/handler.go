package handlers

import (
	"html/template"
	"net/http"

	"github.com/antikuz/demo-avia/internal/models"
	"github.com/antikuz/demo-avia/internal/processors"
	"github.com/antikuz/demo-avia/pkg/logging"
	"github.com/julienschmidt/httprouter"
)

const (
	rootURL = "/"
	searchURL = "/search"
	buyURL = "/buy/:id"
	userURL = "/users/:uuid"
)


type handler struct {
	templates *template.Template
	processor *processors.StorageProcessor
	logger *logging.Logger
}

func NewHandler(templates *template.Template, processor *processors.StorageProcessor, logger *logging.Logger) *handler {
	return &handler{
		templates: templates,
		logger: logger,
		processor: processor,
	}
}

func (h *handler) Register(router *httprouter.Router) {
	router.GET(rootURL, h.GetMain)
	router.GET(buyURL, h.BuyTicket)
	router.POST(searchURL, h.GetFlights)
}

func (h *handler) GetMain(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if err := h.templates.ExecuteTemplate(w, "main.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) GetFlights(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	r.ParseForm()
	postFormValues := r.PostForm
	result := models.SearchResult{
		SearchValues: postFormValues,
		SearchResults: h.processor.List(postFormValues),
	}
	if err := h.templates.ExecuteTemplate(w, "search.html", result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) BuyTicket(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	r.ParseForm()
	postFormValues := r.PostForm
	id := params.ByName("id")
	result := models.BuyFlightID{
		SearchValues: postFormValues,
		SearchResults: h.processor.GetFlight(id),
	}
	if err := h.templates.ExecuteTemplate(w, "buy-ticket.html", result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}