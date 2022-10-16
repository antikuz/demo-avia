package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/antikuz/demo-avia/internal/models"
	"github.com/antikuz/demo-avia/internal/processors"
	"github.com/antikuz/demo-avia/pkg/logging"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

const (
	rootURL = "/"
	searchURL = "/search"
	buyURL = "/buy/:id"
	signinURL = "/signin"
	userProfileURL = "/profile"
	userURL = "/users/:uuid"
)

type handler struct {
	templates *template.Template
	processor *processors.StorageProcessor
	sessions map[string]*models.Session
	logger *logging.Logger
}

func NewHandler(templates *template.Template, processor *processors.StorageProcessor, sessions map[string]*models.Session, logger *logging.Logger) *handler {
	return &handler{
		templates: templates,
		logger: logger,
		sessions: sessions,
		processor: processor,
	}
}

func (h *handler) isAuthorized(w http.ResponseWriter, r *http.Request) bool {
	_, err := r.Cookie("session_token")
	return err == nil
}

func (h *handler) getUser(r *http.Request) string {
	token, err := r.Cookie("session_token")
	if err != nil {
		h.logger.Errorf("Error getUser, due to err: %v", err)
	}
	return h.sessions[token.Value].Username
}

func (h *handler) Register(router *httprouter.Router) {
	router.GET(rootURL, h.GetMain)
	router.GET(buyURL, h.BuyTicket)
	router.POST(searchURL, h.GetFlights)
	router.GET(signinURL, h.SignIn)
	router.POST(signinURL, h.SignIn)
	router.GET(userProfileURL, h.UserProfile)
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
	if !h.isAuthorized(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}
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

func (h *handler) SignIn(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if r.Method == "GET" {
		if err := h.templates.ExecuteTemplate(w, "signin.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		r.ParseForm()
		creds := r.PostForm
		user := h.processor.GetUser(creds["username"][0])
		if user.Password == "" || user.Password != creds["password"][0] {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		sessionToken := uuid.NewString()
		expiresAt := time.Now().Add(6000 * time.Second)

		h.sessions[sessionToken] = &models.Session{
			Username: creds["username"][0],
			Expiry:   expiresAt,
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: expiresAt,
		})

		http.Redirect(w ,r, "/profile", http.StatusSeeOther)
	}

}

func (h *handler) UserProfile(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	if !h.isAuthorized(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}

	user := h.getUser(r)
	if err := h.templates.ExecuteTemplate(w, "profile.html", user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}